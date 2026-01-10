package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/noruj-official/full-stack-go-template/internal/service"
)

// Handler is the base handler with shared utilities.
type Handler struct {
	appName        string
	appLogo        string
	featureService service.FeatureService
}

// NewHandler creates a new base handler.
func NewHandler(appName, appLogo string, featureService service.FeatureService) *Handler {
	return &Handler{
		appName:        appName,
		appLogo:        appLogo,
		featureService: featureService,
	}
}

// GetTheme extracts the theme from the request (cookie or client hint).
// Returns the theme name and whether theme management is enabled.
func (h *Handler) GetTheme(r *http.Request) (string, bool) {
	// Check feature flag first
	enabled, err := h.featureService.IsEnabled(r.Context(), "theme_management")
	if err != nil {
		enabled = false // Default to disabled if check fails (safest)
		// Or defaulting to true? "theme_management" implies disabling it turns it off.
		// If flag is missing, we probably want it enabled by default?
		// Plan said "Disable theme switching if theme_management flag is off".
		// Usually flags start false. So if I want it ON by default, I should have seeded it true.
		// Let's assume false means "No theme management support, just light mode".
	}

	if !enabled {
		return "light", false
	}

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
	return theme, true
}

// GetOAuthEnabled checks if the OAuth feature is enabled.
func (h *Handler) GetOAuthEnabled(r *http.Request) bool {
	enabled, err := h.featureService.IsEnabled(r.Context(), "oauth")
	if err != nil {
		return false // Default to disabled if check fails
	}
	return enabled
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

// getIPAddress extracts the client IP address from the request.
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
