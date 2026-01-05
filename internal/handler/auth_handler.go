// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	*Handler
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(base *Handler, authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		Handler:     base,
		authService: authService,
	}
}

// getDashboardURLForRole returns the appropriate dashboard URL based on user role.
func getDashboardURLForRole(user *domain.User) string {
	if user == nil {
		return "/u/dashboard"
	}
	switch user.Role {
	case domain.RoleSuperAdmin:
		return "/s/dashboard"
	case domain.RoleAdmin:
		return "/a/dashboard"
	default:
		return "/u/dashboard"
	}
}

// SignInPage renders the sign in page.
func (h *AuthHandler) SignInPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to appropriate dashboard
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, getDashboardURLForRole(user), http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Title": "Sign In",
	}
	h.RenderWithUser(w, r, "signin.html", data)
}

// SignIn handles user sign in.
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderSignInError(w, r, "", "Invalid form data")
		return
	}

	input := &domain.LoginInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	user, session, err := h.authService.Login(r.Context(), input)
	if err != nil {
		errMsg := "An error occurred"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if domain.IsInvalidCredentialsError(err) {
			errMsg = "Invalid email or password"
		}
		h.renderSignInError(w, r, input.Email, errMsg)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect based on role
	redirectURL := getDashboardURLForRole(user)

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *AuthHandler) renderSignInError(w http.ResponseWriter, r *http.Request, email, errMsg string) {
	data := map[string]any{
		"Title": "Sign In",
		"Email": email,
		"Error": errMsg,
	}

	if isHTMXRequest(r) {
		h.RenderPartial(w, "signin_form.html", data)
		return
	}

	h.RenderWithUser(w, r, "signin.html", data)
}

// SignupPage renders the signup page.
func (h *AuthHandler) SignupPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to appropriate dashboard
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, getDashboardURLForRole(user), http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Title": "Sign Up",
	}
	h.RenderWithUser(w, r, "signup.html", data)
}

// Signup handles user registration.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderSignupError(w, r, nil, "Invalid form data")
		return
	}

	input := &domain.RegisterInput{
		Email:           r.FormValue("email"),
		Name:            r.FormValue("name"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"),
	}

	user, session, err := h.authService.Register(r.Context(), input)
	if err != nil {
		errMsg := "An error occurred"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if domain.IsConflictError(err) {
			errMsg = "An account with this email already exists"
		}
		h.renderSignupError(w, r, input, errMsg)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect based on role
	redirectURL := getDashboardURLForRole(user)

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *AuthHandler) renderSignupError(w http.ResponseWriter, r *http.Request, input *domain.RegisterInput, errMsg string) {
	data := map[string]any{
		"Title": "Sign Up",
		"Form":  input,
		"Error": errMsg,
	}

	if isHTMXRequest(r) {
		h.RenderPartial(w, "signup_form.html", data)
		return
	}

	h.RenderWithUser(w, r, "signup.html", data)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := middleware.GetSessionIDFromContext(r.Context())
	if sessionID != "" {
		_ = h.authService.Logout(r.Context(), sessionID)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LoginRedirect redirects /login to /signin for backwards compatibility.
func (h *AuthHandler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
}
