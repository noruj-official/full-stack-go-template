// Package handler provides HTTP request handlers.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/noruj-official/full-stack-go-template/internal/config"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/repository/postgres"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/admin"
)

// SystemStats holds realtime system statistics
type SystemStats struct {
	CPUUsage       float64
	RAMUsage       float64
	RAMTotal       float64
	RAMUsedPercent float64
	mu             sync.RWMutex
}

// AuditHandler handles audit log HTTP requests.
type AuditHandler struct {
	*Handler
	auditService service.AuditService
	db           *postgres.DB
	cfg          *config.Config
	stats        *SystemStats
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(base *Handler, auditService service.AuditService, db *postgres.DB, cfg *config.Config) *AuditHandler {
	return &AuditHandler{
		Handler:      base,
		auditService: auditService,
		db:           db,
		cfg:          cfg,
		stats:        &SystemStats{},
	}
}

// StartMonitoring starts the background system monitoring
func (h *AuditHandler) StartMonitoring(ctx context.Context) {
	// Initialize CPU counter to avoid first-call garbage/error
	_, _ = cpu.Percent(0, false)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				v, _ := mem.VirtualMemory()
				// Interval 0 means "calculate since last call"
				// Since we prefer accurate "since last tick" measurement without blocking the goroutine
				c, _ := cpu.Percent(0, false)

				h.stats.mu.Lock()
				if len(c) > 0 {
					h.stats.CPUUsage = c[0]
				}
				if v != nil {
					h.stats.RAMUsage = float64(v.Used) / 1024 / 1024 / 1024
					h.stats.RAMTotal = float64(v.Total) / 1024 / 1024 / 1024
					h.stats.RAMUsedPercent = v.UsedPercent
				}
				h.stats.mu.Unlock()
			}
		}
	}()
}

// AuditLogs renders the audit logs page.
func (h *AuditHandler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit := 20
	offset := (page - 1) * limit

	logs, total, err := h.auditService.GetAuditLogs(r.Context(), limit, offset)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load audit logs")
		return
	}

	// Format logs for display
	formattedLogs := make([]admin.AuditLogItem, 0, len(logs))
	for _, log := range logs {
		ip := ""
		if log.IPAddress != nil {
			ip = *log.IPAddress
		}
		formattedLogs = append(formattedLogs, admin.AuditLogItem{
			ID:           log.ID.String(),
			AdminName:    log.AdminName,
			Action:       string(log.Action),
			ResourceType: log.ResourceType,
			TimeAgo:      formatTimeAgo(log.CreatedAt),
			FullTime:     log.CreatedAt.Format("Jan 02, 2006 at 3:04 PM"),
			IPAddress:    ip,
		})
	}

	totalPages := (total + limit - 1) / limit

	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	props := admin.AuditLogsProps{
		User:        middleware.GetUserFromContext(r.Context()),
		Logs:        formattedLogs,
		Total:       int64(total),
		CurrentPage: page,
		TotalPages:  totalPages,

		Theme:        theme,
		ThemeEnabled: themeEnabled,
		OAuthEnabled: oauthEnabled,
	}

	admin.AuditLogs(props).Render(r.Context(), w)
}

// SystemMetricsJSON returns system metrics as JSON for client-side polling
func (h *AuditHandler) SystemMetricsJSON(w http.ResponseWriter, r *http.Request) {
	h.stats.mu.RLock()
	cpuUsage := h.stats.CPUUsage
	ramPercent := h.stats.RAMUsedPercent
	h.stats.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"cpu":%.1f,"ram":%.1f}`, cpuUsage, ramPercent)
}

// SystemHealth renders the system health monitoring page.
func (h *AuditHandler) SystemHealth(w http.ResponseWriter, r *http.Request) {
	// Fast path for polling: check cached stats immediately
	if isHTMXRequest(r) && r.Header.Get("HX-Target-Component") == "SystemMetricsUpdate" {
		h.stats.mu.RLock()
		props := admin.SystemHealthProps{
			Application: admin.AppHealth{
				CPUUsage:        h.stats.CPUUsage,
				RAMUsagePercent: h.stats.RAMUsedPercent,
			},
		}
		h.stats.mu.RUnlock()
		admin.SystemMetricsUpdate(props).Render(r.Context(), w)
		return
	}

	// Check database health
	dbStatus := "Connected"
	dbError := ""
	if err := h.db.Pool.Ping(r.Context()); err != nil {
		dbStatus = "Error"
		dbError = err.Error()
	}

	// Get database pool stats
	poolStats := h.db.Pool.Stat()

	// Get real Postgres version
	var pgVersion string
	err := h.db.Pool.QueryRow(r.Context(), "SELECT version()").Scan(&pgVersion)
	if err != nil {
		pgVersion = "Unknown"
	}

	// Get system usage stats from cache or fallback
	h.stats.mu.RLock()
	cpuUsage := h.stats.CPUUsage
	ramUsage := h.stats.RAMUsage
	ramTotal := h.stats.RAMTotal
	ramPercent := h.stats.RAMUsedPercent
	h.stats.mu.RUnlock()

	// If stats are empty (start up), fetch once synchronously (calls might block but better than showing 0)
	if cpuUsage == 0 && ramTotal == 0 {
		v, _ := mem.VirtualMemory()
		c, _ := cpu.Percent(time.Second, false)
		if len(c) > 0 {
			cpuUsage = c[0]
		}
		if v != nil {
			ramUsage = float64(v.Used) / 1024 / 1024 / 1024
			ramTotal = float64(v.Total) / 1024 / 1024 / 1024
			ramPercent = v.UsedPercent
		}
	}

	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	props := admin.SystemHealthProps{
		User: middleware.GetUserFromContext(r.Context()),
		Database: admin.DatabaseHealth{
			Status:         dbStatus,
			Error:          dbError,
			Type:           pgVersion, // Using full version string
			MaxConnections: poolStats.MaxConns(),
			IdleConns:      poolStats.IdleConns(),
			AcquiredConns:  poolStats.AcquiredConns(),
			TotalConns:     poolStats.TotalConns(),
		},
		Application: admin.AppHealth{
			Name:            h.cfg.App.Name,
			Environment:     h.cfg.App.Env,
			GoVersion:       runtime.Version(),
			GOOS:            runtime.GOOS,
			GOARCH:          runtime.GOARCH,
			NumGoroutine:    runtime.NumGoroutine(),
			NumCPU:          runtime.NumCPU(),
			MemoryUsage:     fmt.Sprintf("%.2f GB / %.2f GB", ramUsage, ramTotal),
			CPUUsage:        cpuUsage,
			RAMUsagePercent: ramPercent,
		},
		Server: admin.ServerHealth{
			ReadTimeout:  h.cfg.Server.ReadTimeout,
			WriteTimeout: h.cfg.Server.WriteTimeout,
			IdleTimeout:  h.cfg.Server.IdleTimeout,
		},

		Theme:        theme,
		ThemeEnabled: themeEnabled,
		OAuthEnabled: oauthEnabled,
	}

	if isHTMXRequest(r) {
		if r.Header.Get("HX-Target") == "system-charts-container" {
			admin.SystemResources(props).Render(r.Context(), w)
			return
		}
	}

	admin.SystemHealth(props).Render(r.Context(), w)
}
