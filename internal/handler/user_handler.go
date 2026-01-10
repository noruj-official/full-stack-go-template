// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	usersPage "github.com/noruj-official/full-stack-go-template/web/templ/pages/users"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	*Handler
	userService  service.UserService
	auditService service.AuditService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(base *Handler, userService service.UserService, auditService service.AuditService) *UserHandler {
	return &UserHandler{
		Handler:      base,
		userService:  userService,
		auditService: auditService,
	}
}

// List renders the users list page.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.userService.ListUsers(r.Context(), page, 10)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load users")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	showSidebar := true

	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, usersPage.List("Users", "Manage your application users", user, showSidebar, theme, themeEnabled, oauthEnabled, users, total, page, int((total+9)/10)))
}

// Create handles user creation form display and submission.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		user := middleware.GetUserFromContext(r.Context())
		theme, themeEnabled := h.GetTheme(r)
		oauthEnabled := h.GetOAuthEnabled(r)
		h.RenderTempl(w, r, usersPage.Create("Create User", "Add a new user to your application", user, true, theme, themeEnabled, oauthEnabled, nil, ""))
		return
	}

	// Handle POST
	if err := r.ParseForm(); err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	input := &domain.CreateUserInput{
		Email: r.FormValue("email"),
		Name:  r.FormValue("name"),
	}

	user, err := h.userService.CreateUser(r.Context(), input)
	if err != nil {
		if domain.IsValidationError(err) {
			h.renderCreateForm(w, r, input, err.Error())
			return
		}
		if domain.IsConflictError(err) {
			h.renderCreateForm(w, r, input, "A user with this email already exists")
			return
		}
		h.Error(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Log audit for user creation (admin context required)
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, domain.AuditUserCreate, "user", &user.ID, nil, map[string]interface{}{
			"email": user.Email,
			"name":  user.Name,
		}, &ip)
	}

	// For HTMX requests, return the new row
	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "userCreated")
		h.RenderTempl(w, r, usersPage.UserRow(user))
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}

func (h *UserHandler) renderCreateForm(w http.ResponseWriter, r *http.Request, input *domain.CreateUserInput, errMsg string) {
	// If HTMX, render just the form content (UserForm)
	if isHTMXRequest(r) {
		h.RenderTempl(w, r, usersPage.UserForm(nil, input, errMsg))
		return
	}

	// Otherwise render the full create page
	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	h.RenderTempl(w, r, usersPage.Create("Create User", "Add a new user to your application", user, true, theme, themeEnabled, oauthEnabled, input, errMsg))
}

// Edit handles user edit form display and submission.
func (h *UserHandler) Edit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			h.Error(w, r, http.StatusNotFound, "User not found")
			return
		}
		h.Error(w, r, http.StatusInternalServerError, "Failed to load user")
		return
	}

	if r.Method == http.MethodGet {
		currentUser := middleware.GetUserFromContext(r.Context())
		theme, themeEnabled := h.GetTheme(r)
		oauthEnabled := h.GetOAuthEnabled(r)
		h.RenderTempl(w, r, usersPage.Edit("Edit User", "Update user information", currentUser, true, theme, themeEnabled, oauthEnabled, user, ""))
		return
	}

	// Handle POST/PUT
	if err := r.ParseForm(); err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")

	input := &domain.UpdateUserInput{
		Email: &email,
		Name:  &name,
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), id, input)
	if err != nil {
		if domain.IsValidationError(err) {
			h.renderEditForm(w, r, user, err.Error())
			return
		}
		if domain.IsConflictError(err) {
			h.renderEditForm(w, r, user, "A user with this email already exists")
			return
		}
		h.Error(w, r, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Log audit for user update
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, domain.AuditUserUpdate, "user", &updatedUser.ID, map[string]interface{}{
			"email": user.Email,
			"name":  user.Name,
		}, map[string]interface{}{
			"email": updatedUser.Email,
			"name":  updatedUser.Name,
		}, &ip)
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "userUpdated")
		h.RenderTempl(w, r, usersPage.UserRow(updatedUser))
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}

func (h *UserHandler) renderEditForm(w http.ResponseWriter, r *http.Request, targetUser *domain.User, errMsg string) {
	if isHTMXRequest(r) {
		// Render just the form content (UserForm)
		// Note: We need to pass targetUser here essentially as the 'user' for UserForm
		h.RenderTempl(w, r, usersPage.UserForm(targetUser, nil, errMsg))
		return
	}

	currentUser := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	h.RenderTempl(w, r, usersPage.Edit("Edit User", "Update user information", currentUser, true, theme, themeEnabled, oauthEnabled, targetUser, errMsg))
}

// Delete handles user deletion.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		if domain.IsNotFoundError(err) {
			h.Error(w, r, http.StatusNotFound, "User not found")
			return
		}
		h.Error(w, r, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Log audit for user deletion
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, domain.AuditUserDelete, "user", &id, nil, nil, &ip)
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "userDeleted")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}

// UpdateStatus handles user status updates.
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	status := domain.UserStatus(r.FormValue("status"))
	switch status {
	case domain.UserStatusActive, domain.UserStatusSuspended, domain.UserStatusBanned:
		// Valid status
	default:
		h.Error(w, r, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.userService.UpdateStatus(r.Context(), id, status); err != nil {
		if domain.IsNotFoundError(err) {
			h.Error(w, r, http.StatusNotFound, "User not found")
			return
		}
		h.Error(w, r, http.StatusInternalServerError, "Failed to update user status")
		return
	}

	// Log audit for status update
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, domain.AuditUserUpdate, "user", &id, nil, map[string]interface{}{
			"status": status,
		}, &ip)
	}

	user, _ := h.userService.GetUser(r.Context(), id)

	if isHTMXRequest(r) {
		w.Header().Set("HX-Trigger", "userStatusUpdated")
		h.RenderTempl(w, r, usersPage.UserRow(user))
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}
