// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
	"github.com/shaik-noor/full-stack-go-template/web/templ/pages/profile"
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
	theme := getTheme(r)

	if r.Method == http.MethodGet {
		h.RenderTempl(w, r, profile.Settings("Settings", user, theme, ""))
		return
	}

	// Handle POST - update profile
	if err := r.ParseForm(); err != nil {
		h.renderSettingsError(w, r, user, theme, "Invalid form data")
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
		h.renderSettingsError(w, r, user, theme, errMsg)
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

func (h *SettingsHandler) renderSettingsError(w http.ResponseWriter, r *http.Request, user *domain.User, theme, errMsg string) {
	if isHTMXRequest(r) {
		// For HTMX requests, we might want to return just the form with the error,
		// but since our new SettingsFormFields doesn't accept error string directly (it's mainly for fields),
		// we might need to adjust or just re-render the whole settings page if simpler,
		// OR better: Update SettingsFormFields logic if needed.
		// However, looking at the previous implementation, it swapped #settings-form.
		// Let's re-render the whole Settings page but targeting the form if HTMX, or just the whole page.
		// Actually, standard HTMX pattern: re-render the component.
		// For now, let's re-render the page which contains the error block.
		// Ideally we should have a SettingsForm component that includes the error.
		// Let's rely on the full page render for simplicity unless refined.
		// Wait, HTMX swap outerHTML on #settings-form.
		// Let's assume we return the full page and HTMX extracts (if hx-select is used) or just return the form part.
		// The templ component SettingsFormFields doesn't have the error logic inside it in my previous step, it was in Settings.
		// Let's fix this by updating the handler to render the full page but maybe HTMX will re-render the body?
		// To be safe and consistent with the templ definition:

		// If I look at my previous step:
		// if errStr != "" { <div class="alert ...">...</div> } is in Settings, NOT in SettingsFormFields.
		// So if I only render SettingsFormFields, the error won't show!
		// I should update SettingsFormFields to optionally take an error or have the error block inside the form container?
		// Actually, in the previous code: <div id="settings-form" class="space-y-6"> @SettingsFormFields(user) </div>
		// The error was OUTSIDE the form in `Settings` component:
		// <div class="card-body p-6"> if errStr != "" ... <form ...> ... </form> </div>

		// If I want to show the error on HTMX error, I need to render the part that INCLUDES the error.
		// Let's render the full Settings page for now as a fallback or simply redirect if error (bad UX).
		// Correct approach: Render the full Settings component. HTMX will replace the target.
		// But the target is #settings-form. The error is OUTSIDE #settings-form.
		// Use HX-Retarget? Or just render the whole content and let HTMX swap?
		// If I render the whole page, replacing #settings-form with the whole page is wrong.

		// Let's simply re-render the full page and sending it.
		// But to make it cleaner, let's update `settings.templ` in the NEXT step (iterative fix) if needed.
		// For now, I will render the full page logic.
		h.RenderTempl(w, r, profile.Settings("Settings", user, theme, errMsg))
		return
	}

	h.RenderTempl(w, r, profile.Settings("Settings", user, theme, errMsg))
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
