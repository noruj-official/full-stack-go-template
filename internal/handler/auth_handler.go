// Package handler provides HTTP request handlers.
package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
	"github.com/shaik-noor/full-stack-go-template/web/templ/pages/auth"
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

	// Check for query params (e.g. error=unverified&email=...)
	email := r.URL.Query().Get("email")
	msgType := ""
	msg := ""

	if r.URL.Query().Get("error") == "unverified" {
		msgType = "info"
		msg = "Email not verified. A new verification link has been sent to " + email
	}

	if r.URL.Query().Get("success") == "registered" {
		msgType = "success"
		msg = "Account created! Please check your email to verify your account."
	}

	props := auth.SigninPageProps{
		Email:       email,
		Error:       "",
		Message:     msg,
		MessageType: msgType,
		Theme:       h.GetTheme(r),
	}
	auth.SigninPage(props).Render(r.Context(), w)
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

	ip := getIPAddress(r)
	ua := r.UserAgent()

	user, session, err := h.authService.Login(r.Context(), input, ip, ua)
	// Check for email verification error
	if err == domain.ErrEmailNotVerified {
		// Redirect back to sign in with error and email
		// We use a query param 'error=unverified' to trigger the specific message
		redirectURL := "/signin?error=unverified&email=" + input.Email

		if isHTMXRequest(r) {
			w.Header().Set("HX-Redirect", redirectURL)
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	if err != nil {
		errMsg := "An error occurred"
		if domain.IsValidationError(err) {
			errMsg = err.Error()
		} else if domain.IsInvalidCredentialsError(err) {
			errMsg = "Invalid email or password"
		} else {
			// Log the actual error for debugging
			log.Printf("Login error for user %s: %v", input.Email, err)
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
	// ip and ua already captured above
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
	props := auth.SigninPageProps{
		Email: email,
		Error: errMsg,
		Theme: h.GetTheme(r),
	}

	if isHTMXRequest(r) {
		auth.SigninForm(props).Render(r.Context(), w)
		return
	}

	auth.SigninPage(props).Render(r.Context(), w)
}

// SignupPage renders the signup page.
func (h *AuthHandler) SignupPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to appropriate dashboard
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, getDashboardURLForRole(user), http.StatusSeeOther)
		return
	}

	props := auth.SignupPageProps{
		Form:  nil,
		Error: "",
		Theme: h.GetTheme(r),
	}
	auth.SignupPage(props).Render(r.Context(), w)
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

	ip := getIPAddress(r)
	ua := r.UserAgent()

	// Register user
	user, err := h.authService.Register(r.Context(), input, ip, ua)
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

	// Remove session cookie setting logic (we don't auto-login anymore)

	// Log signup activity
	// ip and ua already captured above
	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogin, "User registered", &ip, &ua)

	// Redirect to sign in page with success message
	redirectURL := "/signin?success=registered&email=" + input.Email

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *AuthHandler) renderSignupError(w http.ResponseWriter, r *http.Request, input *domain.RegisterInput, errMsg string) {
	props := auth.SignupPageProps{
		Form:  input,
		Error: errMsg,
		Theme: h.GetTheme(r),
	}

	if isHTMXRequest(r) {
		auth.SignupForm(props).Render(r.Context(), w)
		return
	}

	auth.SignupPage(props).Render(r.Context(), w)
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

// VerifyEmailPage handles email verification via token.
func (h *AuthHandler) VerifyEmailPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var props auth.VerifyEmailPageProps
	props.Theme = h.GetTheme(r)

	if token == "" {
		props.Success = false
		props.Message = "No verification token provided. Please use the link from your email."
		auth.VerifyEmailPage(props).Render(r.Context(), w)
		return
	}

	err := h.authService.VerifyEmail(r.Context(), token)
	if err != nil {
		props.Success = false
		if domain.IsNotFoundError(err) || err == domain.ErrInvalidToken {
			props.Message = "This verification link is invalid or has already been used."
		} else if err == domain.ErrTokenExpired {
			props.Message = "This verification link has expired. Please request a new one."
		} else {
			props.Message = "An error occurred during verification. Please try again later."
		}
		auth.VerifyEmailPage(props).Render(r.Context(), w)
		return
	}

	// Success
	props.Success = true
	props.Message = "Your email has been successfully verified! You can now sign in to your account."
	auth.VerifyEmailPage(props).Render(r.Context(), w)
}
