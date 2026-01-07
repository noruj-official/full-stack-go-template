// Package handler provides HTTP request handlers.
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

// Handler is the base handler with shared utilities.
type Handler struct {
	appName string
	appLogo string
}

// NewHandler creates a new base handler.
func NewHandler(appName, appLogo string) *Handler {
	return &Handler{
		appName: appName,
		appLogo: appLogo,
	}
}

// GetTheme extracts the theme from the request (cookie or client hint).
func (h *Handler) GetTheme(r *http.Request) string {
	theme := "light"
	if c, err := r.Cookie("theme"); err == nil && c.Value != "" {
		if c.Value == "dark" {
			theme = "dark"
		} else if c.Value == "light" {
			theme = "light"
		}
	} else {
		// Optional: use client hint if sent
		if v := r.Header.Get("Sec-CH-Prefers-Color-Scheme"); v == "dark" {
			theme = "dark"
		}
	}
	return theme
}

// JSON sends a JSON response.
func (h *Handler) JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// Error sends an error response, automatically detecting HTMX requests.
func (h *Handler) Error(w http.ResponseWriter, r *http.Request, status int, message string) {
	if isHTMXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		w.Write([]byte(`<div class="text-red-500">` + message + `</div>`))
		return
	}

	http.Error(w, message, status)
}

// isHTMXRequest checks if the request is from HTMX.
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func isHTMXBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

// RenderTempl renders a templ component.
func (h *Handler) RenderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering templ component: %v", err)
		h.Error(w, r, http.StatusInternalServerError, "Error rendering template")
	}
}
