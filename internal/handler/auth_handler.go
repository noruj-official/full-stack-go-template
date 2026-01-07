// Package handler provides HTTP request handlers.
package handler

import (
	"net/http"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
	"github.com/shaik-noor/full-stack-go-template/web/templ/pages"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	*Handler
	authService     service.AuthService
	activityService service.ActivityService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(base *Handler, authService service.AuthService, activityService service.ActivityService) *AuthHandler {
	return &AuthHandler{
		Handler:         base,
		authService:     authService,
		activityService: activityService,
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

	props := pages.SigninPageProps{
		Email: "",
		Error: "",
		Theme: h.GetTheme(r),
	}
	pages.SigninPage(props).Render(r.Context(), w)
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

	// Log login activity
	ip := getIPAddress(r)
	ua := r.UserAgent()
	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogin, "User signed in", &ip, &ua)

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
	props := pages.SigninPageProps{
		Email: email,
		Error: errMsg,
		Theme: h.GetTheme(r),
	}

	if isHTMXRequest(r) {
		pages.SigninForm(props).Render(r.Context(), w)
		return
	}

	pages.SigninPage(props).Render(r.Context(), w)
}

// SignupPage renders the signup page.
func (h *AuthHandler) SignupPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to appropriate dashboard
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, getDashboardURLForRole(user), http.StatusSeeOther)
		return
	}

	props := pages.SignupPageProps{
		Form:  nil,
		Error: "",
		Theme: h.GetTheme(r),
	}
	pages.SignupPage(props).Render(r.Context(), w)
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

	// Log signup as login activity
	ip := getIPAddress(r)
	ua := r.UserAgent()
	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogin, "User registered and signed in", &ip, &ua)

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
	props := pages.SignupPageProps{
		Form:  input,
		Error: errMsg,
		Theme: h.GetTheme(r),
	}

	if isHTMXRequest(r) {
		pages.SignupForm(props).Render(r.Context(), w)
		return
	}

	pages.SignupPage(props).Render(r.Context(), w)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
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

	// Log logout activity when user context is present
	if user != nil {
		ip := getIPAddress(r)
		ua := r.UserAgent()
		_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogout, "User logged out", &ip, &ua)
	}

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
