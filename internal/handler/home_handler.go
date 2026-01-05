// Package handler provides HTTP request handlers.
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-starter/internal/domain"
	"github.com/go-starter/internal/middleware"
	"github.com/go-starter/internal/repository/postgres"
)

// HomeHandler handles home page requests.
type HomeHandler struct {
	*Handler
	db *postgres.DB
}

// NewHomeHandler creates a new home handler.
func NewHomeHandler(base *Handler, db *postgres.DB) *HomeHandler {
	return &HomeHandler{
		Handler: base,
		db:      db,
	}
}

// Index renders the home page.
func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":       "Go Starter",
		"Description": "A professional full-stack Go application",
	}
	h.RenderWithUser(w, r, "home.html", data)
}

// DashboardRedirect redirects to the appropriate dashboard based on user role.
func (h *HomeHandler) DashboardRedirect(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	redirectURL := "/u/dashboard"
	if user != nil {
		switch user.Role {
		case domain.RoleSuperAdmin:
			redirectURL = "/s/dashboard"
		case domain.RoleAdmin:
			redirectURL = "/a/dashboard"
		}
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// UserDashboard renders the user dashboard page.
func (h *HomeHandler) UserDashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	data := map[string]any{
		"Title":       "Dashboard",
		"ShowSidebar": true,
		"User":        user,
	}
	h.RenderWithUser(w, r, "user_dashboard.html", data)
}

// AdminDashboard renders the admin dashboard page.
func (h *HomeHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	userRepo := postgres.NewUserRepository(h.db)

	// Get total user count
	userCount, err := userRepo.Count(r.Context())
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load user stats")
		return
	}

	// Get recent users (last 5)
	users, err := userRepo.List(r.Context(), 5, 0)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load recent activity")
		return
	}

	// Prepare lightweight view model for recent activity
	recent := make([]map[string]string, 0, len(users))
	for _, u := range users {
		ago := humanizeDuration(time.Since(u.CreatedAt))
		recent = append(recent, map[string]string{
			"Name":       u.Name,
			"Email":      u.Email,
			"CreatedAgo": ago,
		})
	}

	data := map[string]any{
		"Title":       "Admin Dashboard",
		"UserCount":   userCount,
		"RecentUsers": recent,
		"ShowSidebar": true,
	}
	h.RenderWithUser(w, r, "admin_dashboard.html", data)
}

// SuperAdminDashboard renders the super admin dashboard page.
func (h *HomeHandler) SuperAdminDashboard(w http.ResponseWriter, r *http.Request) {
	userRepo := postgres.NewUserRepository(h.db)

	// Get total user count
	userCount, err := userRepo.Count(r.Context())
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load user stats")
		return
	}

	// Get admin count
	allUsers, err := userRepo.List(r.Context(), 1000, 0)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load admin stats")
		return
	}

	adminCount := 0
	for _, u := range allUsers {
		if u.Role == domain.RoleAdmin || u.Role == domain.RoleSuperAdmin {
			adminCount++
		}
	}

	// Get recent users (last 5)
	users, err := userRepo.List(r.Context(), 5, 0)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load recent activity")
		return
	}

	// Prepare lightweight view model for recent activity
	recent := make([]map[string]string, 0, len(users))
	for _, u := range users {
		ago := humanizeDuration(time.Since(u.CreatedAt))
		recent = append(recent, map[string]string{
			"Name":       u.Name,
			"Email":      u.Email,
			"Role":       string(u.Role),
			"CreatedAgo": ago,
		})
	}

	data := map[string]any{
		"Title":       "Super Admin Dashboard",
		"UserCount":   userCount,
		"AdminCount":  adminCount,
		"RecentUsers": recent,
		"ShowSidebar": true,
	}
	h.RenderWithUser(w, r, "superadmin_dashboard.html", data)
}

// HealthCheck returns the health status of the application.
func (h *HomeHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	statusCode := http.StatusOK

	if err := h.db.Health(r.Context()); err != nil {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	h.JSON(w, statusCode, map[string]string{
		"status": status,
	})
}

// humanizeDuration returns a short relative time like "2m ago" or "1h ago".
func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return formatUnit(int(d.Minutes()), "m")
	}
	if d < 24*time.Hour {
		return formatUnit(int(d.Hours()), "h")
	}
	return formatUnit(int(d.Hours()/24), "d")
}

func formatUnit(value int, unit string) string {
	if value <= 1 {
		return "1" + unit + " ago"
	}
	return fmt.Sprintf("%d%s ago", value, unit)
}
