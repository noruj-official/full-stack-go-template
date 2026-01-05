// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
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

	data := map[string]any{
		"Title":       "Users",
		"Users":       users,
		"Total":       total,
		"CurrentPage": page,
		"TotalPages":  (total + 9) / 10,
		"ShowSidebar": true,
	}

	if isHTMXRequest(r) && !isHTMXBoosted(r) {
		h.RenderPartialWithUser(w, r, "user_list.html", data)
		return
	}

	h.RenderWithUser(w, r, "list.html", data)
}

// Create handles user creation form display and submission.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.RenderWithUser(w, r, "create.html", map[string]any{
			"Title":       "Create User",
			"ShowSidebar": true,
		})
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
		h.RenderPartialWithUser(w, r, "user_row.html", user)
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}

func (h *UserHandler) renderCreateForm(w http.ResponseWriter, r *http.Request, input *domain.CreateUserInput, errMsg string) {
	data := map[string]any{
		"Title": "Create User",
		"Form":  input,
		"Error": errMsg,
	}

	if isHTMXRequest(r) {
		h.RenderPartialWithUser(w, r, "user_form.html", data)
		return
	}

	h.RenderWithUser(w, r, "create.html", data)
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
		h.RenderWithUser(w, r, "edit.html", map[string]any{
			"Title":       "Edit User",
			"User":        user,
			"ShowSidebar": true,
		})
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
		h.RenderPartialWithUser(w, r, "user_row.html", updatedUser)
		return
	}

	http.Redirect(w, r, "/a/users", http.StatusSeeOther)
}

func (h *UserHandler) renderEditForm(w http.ResponseWriter, r *http.Request, user *domain.User, errMsg string) {
	data := map[string]any{
		"Title": "Edit User",
		"User":  user,
		"Error": errMsg,
	}

	if isHTMXRequest(r) {
		h.RenderPartialWithUser(w, r, "user_form.html", data)
		return
	}

	h.RenderWithUser(w, r, "edit.html", data)
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
