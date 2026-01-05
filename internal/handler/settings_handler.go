// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
)

// SettingsHandler handles user settings HTTP requests.
type SettingsHandler struct {
	*Handler
	userService     service.UserService
	activityService service.ActivityService
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(base *Handler, userService service.UserService, activityService service.ActivityService) *SettingsHandler {
	return &SettingsHandler{
		Handler:         base,
		userService:     userService,
		activityService: activityService,
	}
}

// Settings renders the user settings page.
func (h *SettingsHandler) Settings(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		data := map[string]any{
			"Title":       "Settings",
			"ShowSidebar": true,
		}
		h.RenderWithUser(w, r, "user_settings.html", data)
		return
	}

	// Handle POST - update profile
	if err := r.ParseForm(); err != nil {
		h.renderSettingsError(w, r, "Invalid form data")
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	input := &domain.UpdateUserInput{
		Email: &email,
		Name:  &name,
	}

	_, err := h.userService.UpdateUser(r.Context(), user.ID, input)
	if err != nil {
		errMsg := "Failed to update profile"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if domain.IsConflictError(err) {
			errMsg = "This email is already in use"
		}
		h.renderSettingsError(w, r, errMsg)
		return
	}

	// Log the activity
	ipAddr := getIPAddress(r)
	userAgent := r.UserAgent()
	_ = h.activityService.LogActivity(
		r.Context(),
		user.ID,
		domain.ActivityProfileUpdate,
		"Updated profile information",
		&ipAddr,
		&userAgent,
	)

	// Success response
	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "profileUpdated")
		h.RenderPartialWithUser(w, r, "settings_success.html", map[string]any{
			"Message": "Profile updated successfully",
		})
		return
	}

	http.Redirect(w, r, "/u/settings", http.StatusSeeOther)
}

func (h *SettingsHandler) renderSettingsError(w http.ResponseWriter, r *http.Request, errMsg string) {
	data := map[string]any{
		"Title": "Settings",
		"Error": errMsg,
	}

	if isHTMXRequest(r) {
		h.RenderPartialWithUser(w, r, "settings_form.html", data)
		return
	}

	h.RenderWithUser(w, r, "user_settings.html", data)
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
