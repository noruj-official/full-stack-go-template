// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"runtime"
	"strconv"

	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/repository/postgres"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
)

// AuditHandler handles audit log HTTP requests.
type AuditHandler struct {
	*Handler
	auditService service.AuditService
	db           *postgres.DB
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(base *Handler, auditService service.AuditService, db *postgres.DB) *AuditHandler {
	return &AuditHandler{
		Handler:      base,
		auditService: auditService,
		db:           db,
	}
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
	formattedLogs := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		formattedLogs = append(formattedLogs, map[string]interface{}{
			"ID":           log.ID.String(),
			"AdminName":    log.AdminName,
			"Action":       log.Action,
			"ResourceType": log.ResourceType,
			"TimeAgo":      formatTimeAgo(log.CreatedAt),
			"FullTime":     log.CreatedAt.Format("Jan 02, 2006 at 3:04 PM"),
			"IPAddress":    log.IPAddress,
		})
	}

	totalPages := (total + limit - 1) / limit

	data := map[string]any{
		"Title":       "Audit Logs",
		"Logs":        formattedLogs,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Total":       total,
		"ShowSidebar": true,
	}

	h.RenderWithUser(w, r, "audit_logs.html", data)
}

// SystemHealth renders the system health monitoring page.
func (h *AuditHandler) SystemHealth(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Check database health
	dbStatus := "Connected"
	dbError := ""
	if err := h.db.Pool.Ping(r.Context()); err != nil {
		dbStatus = "Error"
		dbError = err.Error()
	}

	// Get database pool stats
	poolStats := h.db.Pool.Stat()

	// Get system uptime (calculate from start time - would need to be tracked)
	// For now, we'll show the current server status

	// Get environment from config or default
	environment := "Development"
	if appEnv := r.Context().Value("app_env"); appEnv != nil {
		environment = appEnv.(string)
	}

	data := map[string]any{
		"Title":       "System Health",
		"User":        user,
		"ShowSidebar": true,
		"Database": map[string]any{
			"Status":         dbStatus,
			"Error":          dbError,
			"Type":           "PostgreSQL",
			"MaxConnections": poolStats.MaxConns(),
			"IdleConns":      poolStats.IdleConns(),
			"AcquiredConns":  poolStats.AcquiredConns(),
			"TotalConns":     poolStats.TotalConns(),
		},
		"Application": map[string]any{
			"Name":         "Full Stack Go Template",
			"Environment":  environment,
			"GoVersion":    runtime.Version(),
			"GOOS":         runtime.GOOS,
			"GOARCH":       runtime.GOARCH,
			"NumGoroutine": runtime.NumGoroutine(),
			"NumCPU":       runtime.NumCPU(),
		},
		"Server": map[string]any{
			"ReadTimeout":  "15s",
			"WriteTimeout": "15s",
			"IdleTimeout":  "60s",
		},
	}

	h.RenderWithUser(w, r, "system_health.html", data)
}
