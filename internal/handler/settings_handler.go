// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"

	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/profile"
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
	theme, themeEnabled := h.GetTheme(r)

	hasPassword := user.PasswordHash != ""

	if r.Method == http.MethodGet {
		oauthEnabled := h.GetOAuthEnabled(r)
		h.RenderTempl(w, r, profile.Settings("Settings", user, theme, themeEnabled, oauthEnabled, "", hasPassword))
		return
	}

	// Handle POST - update profile
	if err := r.ParseForm(); err != nil {
		h.renderSettingsError(w, r, user, theme, themeEnabled, "Invalid form data")
		return
	}

	name := r.FormValue("name")
	// Email is read-only in settings

	input := &domain.UpdateUserInput{
		Name: &name,
	}

	_, err := h.userService.UpdateUser(r.Context(), user.ID, input)
	if err != nil {
		errMsg := "Failed to update profile"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if domain.IsConflictError(err) {
			errMsg = "This email is already in use"
		}
		h.renderSettingsError(w, r, user, theme, themeEnabled, errMsg)
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
		h.RenderTempl(w, r, profile.SettingsSuccess("Profile updated successfully"))
		return
	}

	http.Redirect(w, r, "/u/settings", http.StatusSeeOther)
}

func (h *SettingsHandler) renderSettingsError(w http.ResponseWriter, r *http.Request, user *domain.User, theme string, themeEnabled bool, errMsg string) {
	// Re-render the settings page with the error message
	// For HTMX requests, this will swap the target (e.g., body or form container)
	// For full page loads, this renders the full HTML
	hasPassword := user.PasswordHash != ""
	oauthEnabled := h.GetOAuthEnabled(r)
	h.RenderTempl(w, r, profile.Settings("Settings", user, theme, themeEnabled, oauthEnabled, errMsg, hasPassword))
}

// UpdatePassword handles password update requests.
func (h *SettingsHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	hasPassword := user.PasswordHash != ""

	if err := r.ParseForm(); err != nil {
		h.RenderTempl(w, r, profile.PasswordUpdateForm("Invalid form data", hasPassword))
		return
	}

	input := &domain.UpdatePasswordInput{
		CurrentPassword: r.FormValue("current_password"),
		NewPassword:     r.FormValue("new_password"),
		ConfirmPassword: r.FormValue("confirm_password"),
	}

	if err := h.userService.UpdatePassword(r.Context(), user.ID, input); err != nil {
		errMsg := "Failed to update password"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if err == domain.ErrInvalidCredentials {
			errMsg = "Invalid current password"
		}
		h.RenderTempl(w, r, profile.PasswordUpdateForm(errMsg, hasPassword))
		return
	}

	// Log the activity
	ipAddr := getIPAddress(r)
	userAgent := r.UserAgent()
	_ = h.activityService.LogActivity(
		r.Context(),
		user.ID,
		domain.ActivityPasswordChange,
		"Updated password",
		&ipAddr,
		&userAgent,
	)

	h.RenderTempl(w, r, profile.SettingsSuccess("Password updated successfully"))
}
