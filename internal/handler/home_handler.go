// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"

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
	h.Render(w, "home.html", data)
}

// Dashboard renders the dashboard page.
func (h *HomeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Dashboard",
	}
	h.Render(w, "dashboard.html", data)
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
