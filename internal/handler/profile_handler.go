// Package handler provides HTTP request handlers.
package handler

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/profile"
)

// ProfileHandler handles user profile HTTP requests.
type ProfileHandler struct {
	*Handler
	userService     service.UserService
	activityService service.ActivityService
	mediaService    *service.MediaService
}

// NewProfileHandler creates a new profile handler.
func NewProfileHandler(base *Handler, userService service.UserService, activityService service.ActivityService, mediaService *service.MediaService) *ProfileHandler {
	return &ProfileHandler{
		Handler:         base,
		userService:     userService,
		activityService: activityService,
		mediaService:    mediaService,
	}
}

// ProfilePage renders the user profile page.
func (h *ProfileHandler) ProfilePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	props := profile.UserProfileProps{
		User:         user,
		Theme:        theme,
		ThemeEnabled: themeEnabled,
		OAuthEnabled: oauthEnabled,
	}

	profile.UserProfile(props).Render(r.Context(), w)
}

// UpdateProfile handles profile information updates (name, email).
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		h.Error(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderProfileError(w, r, "Invalid form data")
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
		h.renderProfileError(w, r, errMsg)
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
		profile.ProfileSuccess("Profile updated successfully").Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/u/profile", http.StatusSeeOther)
}

// UploadProfileImage handles profile image uploads.
func (h *ProfileHandler) UploadProfileImage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		h.Error(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.renderProfileError(w, r, "Failed to parse upload form")
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("profile_image")
	if err != nil {
		h.renderProfileError(w, r, "No image file provided")
		return
	}
	defer file.Close()

	// Read file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		h.renderProfileError(w, r, "Failed to read image file")
		return
	}

	// Create media input
	mediaInput := domain.CreateMediaInput{
		UserID:      &user.ID,
		Filename:    header.Filename,
		Data:        imageData,
		ContentType: header.Header.Get("Content-Type"),
		SizeBytes:   len(imageData),
		AltText:     "Profile Image",
	}

	// Upload to MediaService
	media, err := h.mediaService.Upload(r.Context(), mediaInput)
	if err != nil {
		h.renderProfileError(w, r, "Failed to upload profile image: "+err.Error())
		return
	}

	// Update user profile with MediaID
	updateInput := &domain.UpdateUserInput{
		ProfileMediaID: &media.ID,
	}
	_, err = h.userService.UpdateUser(r.Context(), user.ID, updateInput)
	if err != nil {
		h.renderProfileError(w, r, "Failed to update profile image reference: "+err.Error())
		return
	}

	// Log the activity
	ipAddr := getIPAddress(r)
	userAgent := r.UserAgent()
	_ = h.activityService.LogActivity(
		r.Context(),
		user.ID,
		domain.ActivityProfileUpdate,
		"Updated profile image",
		&ipAddr,
		&userAgent,
	)

	// Success response
	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "profileImageUpdated")
		// For image upload, we might want to return the success message or just empty/status
		// The original code rendered profile_success.html
		profile.ProfileSuccess("Profile image updated successfully").Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/u/profile", http.StatusSeeOther)
}

// GetMyProfileImage retrieves the current user's profile image.
func (h *ProfileHandler) GetMyProfileImage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user.ProfileMediaID == nil {
		http.NotFound(w, r)
		return
	}

	h.serveMedia(w, r, *user.ProfileMediaID)
}

// GetUserProfileImage retrieves any user's profile image by ID.
func (h *ProfileHandler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from path
	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch user to get ProfileMediaID
	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil || user == nil || user.ProfileMediaID == nil {
		http.NotFound(w, r)
		return
	}

	h.serveMedia(w, r, *user.ProfileMediaID)
}

// serveMedia is a helper function to serve media content.
func (h *ProfileHandler) serveMedia(w http.ResponseWriter, r *http.Request, mediaID uuid.UUID) {
	media, err := h.mediaService.GetByID(r.Context(), mediaID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set headers for content
	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Media IDs are immutable (mostly), safe to cache
	// w.Header().Set("Vary", "Cookie") // Not needed if serving by immutable ID associated with user

	// Write image data
	w.WriteHeader(http.StatusOK)
	w.Write(media.Data)
}

func (h *ProfileHandler) renderProfileError(w http.ResponseWriter, r *http.Request, errMsg string) {
	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	props := profile.UserProfileProps{
		User:         user,
		Error:        errMsg,
		Theme:        theme,
		ThemeEnabled: themeEnabled,
		OAuthEnabled: oauthEnabled,
	}

	if isHTMXRequest(r) {
		profile.ProfileForm(props).Render(r.Context(), w)
		return
	}

	profile.UserProfile(props).Render(r.Context(), w)
}
