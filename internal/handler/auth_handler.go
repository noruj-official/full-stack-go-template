// Package handler provides HTTP request handlers.
package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/auth"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	*Handler
	authService     service.AuthService
	userService     service.UserService
	activityService service.ActivityService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(base *Handler, authService service.AuthService, userService service.UserService, activityService service.ActivityService) *AuthHandler {
	return &AuthHandler{
		Handler:         base,
		authService:     authService,
		userService:     userService,
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

	if r.URL.Query().Get("error") == "account_suspended" {
		msgType = "error"
		msg = "Your account has been suspended. Please contact support for assistance."
	}

	if r.URL.Query().Get("success") == "registered" {
		msgType = "success"
		msg = "Account created! Please check your email to verify your account."
	}

	if r.URL.Query().Get("success") == "email_sent" {
		msgType = "success"
		msg = fmt.Sprintf("Magic Sign-in link sent to %s. Please check your inbox.", email)
	}

	theme, themeEnabled := h.GetTheme(r)

	// Check email auth feature
	emailAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
	if err != nil {
		emailAuthEnabled = true // Default to true if check fails
	}

	// Check password auth feature
	emailPasswordAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)
	if err != nil {
		emailPasswordAuthEnabled = true // Default to true if check fails
	}

	oauthEnabled, _ := h.authService.ListEnabledProviders(r.Context())

	props := auth.SigninPageProps{
		Email:                    email,
		Error:                    "",
		Message:                  msg,
		MessageType:              msgType,
		Theme:                    theme,
		ThemeEnabled:             themeEnabled,
		EmailAuthEnabled:         emailAuthEnabled,
		EmailPasswordAuthEnabled: emailPasswordAuthEnabled,
		OAuthEnabled:             oauthEnabled,
	}
	auth.SigninPage(props).Render(r.Context(), w)
}

// SignIn handles user sign in.
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	// Check feature flag for password auth
	enabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)
	if err == nil && !enabled {
		h.Error(w, r, http.StatusForbidden, "Email/Password sign in is currently disabled")
		return
	}

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
	theme, themeEnabled := h.GetTheme(r)

	// Check feature flags
	emailAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
	if err != nil {
		emailAuthEnabled = true
	}

	emailPasswordAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)
	if err != nil {
		emailPasswordAuthEnabled = true
	}

	props := auth.SigninPageProps{
		Email:                    email,
		Error:                    errMsg,
		Theme:                    theme,
		ThemeEnabled:             themeEnabled,
		EmailAuthEnabled:         emailAuthEnabled,
		EmailPasswordAuthEnabled: emailPasswordAuthEnabled,
		OAuthEnabled:             nil,
	}

	// Fetch oauth providers even in error for consistent UI
	if oauthEnabled, err := h.authService.ListEnabledProviders(r.Context()); err == nil {
		props.OAuthEnabled = oauthEnabled
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

	theme, themeEnabled := h.GetTheme(r)

	// Check feature flags
	emailAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
	if err != nil {
		emailAuthEnabled = true
	}

	emailPasswordAuthEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)
	if err != nil {
		emailPasswordAuthEnabled = true
	}

	props := auth.SignupPageProps{
		Form:                     nil,
		Error:                    "",
		Theme:                    theme,
		ThemeEnabled:             themeEnabled,
		EmailAuthEnabled:         emailAuthEnabled,
		EmailPasswordAuthEnabled: emailPasswordAuthEnabled,
		OAuthEnabled:             nil,
	}

	if oauthEnabled, err := h.authService.ListEnabledProviders(r.Context()); err == nil {
		props.OAuthEnabled = oauthEnabled
	}

	auth.SignupPage(props).Render(r.Context(), w)
}

// Signup handles user registration.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	// Check feature flag
	enabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)
	if err == nil && !enabled {
		h.Error(w, r, http.StatusForbidden, "Sign up is currently disabled")
		return
	}

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
	theme, themeEnabled := h.GetTheme(r)
	props := auth.SignupPageProps{
		Form:  input,
		Error: errMsg,
		Theme: theme,

		ThemeEnabled:     themeEnabled,
		EmailAuthEnabled: true,
		OAuthEnabled:     nil,
	}
	if oauthEnabled, err := h.authService.ListEnabledProviders(r.Context()); err == nil {
		props.OAuthEnabled = oauthEnabled
	}

	if isHTMXRequest(r) {
		auth.SignupForm(props).Render(r.Context(), w)
		return
	}

	auth.SignupPage(props).Render(r.Context(), w)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/signin")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

// SignOutAllDevices handles invalidating all user sessions.
func (h *AuthHandler) SignOutAllDevices(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if err := h.authService.SignOutAllDevices(r.Context(), user.ID); err != nil {
		log.Printf("Failed to sign out all devices for user %s: %v", user.ID, err)
		// Continue to logout current session anyway
	}

	// Log activity
	ip := getIPAddress(r)
	ua := r.UserAgent()
	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogout, "User signed out all devices", &ip, &ua)

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/signin?success=signed_out_all")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signin?success=signed_out_all", http.StatusSeeOther)
}

// ForgotPasswordPage renders the forgot password page.
func (h *AuthHandler) ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	theme, themeEnabled := h.GetTheme(r)
	h.RenderTempl(w, r, auth.ForgotPassword(theme, themeEnabled))
}

// ForgotPassword handles the forgot password request.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	emailStr := r.FormValue("email")

	if err := h.authService.RequestPasswordReset(r.Context(), emailStr); err != nil {
		// Log error but don't reveal failure to user (security best practice)
		log.Printf("Password reset request failed: %v", err)
	}

	// Always show success message to prevent user enumeration
	theme, themeEnabled := h.GetTheme(r)
	if isHTMXRequest(r) {
		h.RenderTempl(w, r, auth.ForgotPasswordSuccessContent(theme, themeEnabled))
		return
	}
	h.RenderTempl(w, r, auth.ForgotPasswordSuccess(theme, themeEnabled))
}

// ResetPasswordPage renders the reset password page.
func (h *AuthHandler) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	theme, themeEnabled := h.GetTheme(r)
	h.RenderTempl(w, r, auth.ResetPassword(token, theme, themeEnabled))
}

// ResetPassword handles the password reset.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		// In a real app, render the form again with error
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	if err := h.authService.ResetPassword(r.Context(), token, password); err != nil {
		http.Error(w, "Failed to reset password: "+err.Error(), http.StatusBadRequest)
		return
	}

	theme, themeEnabled := h.GetTheme(r)
	if isHTMXRequest(r) {
		h.RenderTempl(w, r, auth.ResetPasswordSuccessContent(theme, themeEnabled))
		return
	}

	h.RenderTempl(w, r, auth.ResetPasswordSuccess(theme, themeEnabled))
}

// LoginRedirect redirects /login to /signin for backwards compatibility.
func (h *AuthHandler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
}

// VerifyEmailPage handles email verification via token.
func (h *AuthHandler) VerifyEmailPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var props auth.VerifyEmailPageProps
	theme, themeEnabled := h.GetTheme(r)
	props.Theme = theme
	props.ThemeEnabled = themeEnabled

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

// HandleOAuthLogin initiates the OAuth login flow.
func (h *AuthHandler) HandleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = r.PathValue("provider")
	}

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	// Generate state (random string) to prevent CSRF
	// In a real app, store this in session or cookie
	state := "random_state_string" // TODO: Implement proper state handling

	url, err := h.authService.GetOAuthLoginURL(r.Context(), domain.OAuthProviderType(provider), state)
	if err != nil {
		fmt.Printf("DEBUG: HandleOAuthLogin failed: %v\n", err)
		log.Printf("Failed to get oauth login url: %v", err)
		http.Redirect(w, r, "/signin?error=oauth_failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

// HandleOAuthCallback handles the OAuth callback.
func (h *AuthHandler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = r.PathValue("provider")
	}

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Redirect(w, r, "/signin?error=oauth_failed", http.StatusSeeOther)
		return
	}

	// Verify state...

	ip := getIPAddress(r)
	ua := r.UserAgent()

	user, session, err := h.authService.LoginWithOAuth(r.Context(), domain.OAuthProviderType(provider), code, state, ip, ua)
	if err != nil {
		log.Printf("OAuth login failed for %s: %v", provider, err)
		http.Redirect(w, r, "/signin?error=oauth_failed", http.StatusSeeOther)
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

	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogin, fmt.Sprintf("User signed in with %s", provider), &ip, &ua)

	// Redirect
	http.Redirect(w, r, getDashboardURLForRole(user), http.StatusSeeOther)
}

// HandleEmailAuthRequest handles the request to sign in/up with email.
func (h *AuthHandler) HandleEmailAuthRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check feature flag
	enabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
	if err == nil && !enabled {
		h.Error(w, r, http.StatusForbidden, "Email authentication is disabled")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		// Render error on signin page
		h.renderSignInError(w, r, "", "Email is required")
		return
	}

	// 1. Generate Token
	token, err := h.authService.GenerateEmailAuthToken(email, domain.TokenPurposeEmailAuth)
	if err != nil {
		log.Printf("Failed to generate email auth token: %v", err)
		h.renderSignInError(w, r, email, "An error occurred")
		return
	}

	// 2. Send Email
	// Use goroutine to not block response
	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := h.authService.SendEmailAuthLink(sendCtx, email, token); err != nil {
			log.Printf("Failed to send email auth link to %s: %v", email, err)
		}
	}()

	// 3. Render Success Page or Message
	// We should probably redirect to a "Check your email" page or show a success message on the signin page.
	// Let's redirect to signin with a success message query param.
	redirectURL := "/signin?success=email_sent&email=" + email

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// HandleEmailAuthVerify handles the verification of the email auth token.
func (h *AuthHandler) HandleEmailAuthVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/signin?error=invalid_token", http.StatusSeeOther)
		return
	}

	// Check feature flag
	enabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
	if err == nil && !enabled {
		h.Error(w, r, http.StatusForbidden, "Email authentication is disabled")
		return
	}

	ip := getIPAddress(r)
	ua := r.UserAgent()

	// Verify and Login
	user, session, err := h.authService.LoginWithEmailToken(r.Context(), token, ip, ua)
	if err != nil {
		log.Printf("Email auth login failed: %v", err)
		if err == domain.ErrInvalidToken || err == domain.ErrTokenExpired {
			http.Redirect(w, r, "/signin?error=invalid_token", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/signin?error=server_error", http.StatusSeeOther)
		}
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

	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityLogin, "User signed in via email", &ip, &ua)

	// Check if user has a name (for signup flow)
	if user.Name == "" {
		redirectURL := "/auth/complete-profile"
		if isHTMXRequest(r) {
			w.Header().Set("HX-Redirect", redirectURL)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Redirect to dashboard
	redirectURL := getDashboardURLForRole(user)
	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// CompleteProfilePage renders the profile completion page.
func (h *AuthHandler) CompleteProfilePage(w http.ResponseWriter, r *http.Request) {
	theme, themeEnabled := h.GetTheme(r)
	props := auth.CompleteProfilePageProps{
		Theme:        theme,
		ThemeEnabled: themeEnabled,
	}
	h.RenderTempl(w, r, auth.CompleteProfilePage(props))
}

// CompleteProfile handles the profile completion submission.
func (h *AuthHandler) CompleteProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		theme, themeEnabled := h.GetTheme(r)
		props := auth.CompleteProfilePageProps{
			Error:        "Name is required",
			Theme:        theme,
			ThemeEnabled: themeEnabled,
		}
		if isHTMXRequest(r) {
			h.RenderTempl(w, r, auth.CompleteProfileForm(props))
			return
		}
		h.RenderTempl(w, r, auth.CompleteProfilePage(props))
		return
	}

	// Update user name
	updateInput := &domain.UpdateUserInput{
		Name: &name,
	}

	_, err := h.userService.UpdateUser(r.Context(), user.ID, updateInput)
	if err != nil {
		log.Printf("Failed to update user profile: %v", err)
		theme, themeEnabled := h.GetTheme(r)
		props := auth.CompleteProfilePageProps{
			Error:        "Failed to update profile",
			Theme:        theme,
			ThemeEnabled: themeEnabled,
		}
		if isHTMXRequest(r) {
			h.RenderTempl(w, r, auth.CompleteProfileForm(props))
			return
		}
		h.RenderTempl(w, r, auth.CompleteProfilePage(props))
		return
	}

	// Log activity
	ip := getIPAddress(r)
	ua := r.UserAgent()
	_ = h.activityService.LogActivity(r.Context(), user.ID, domain.ActivityProfileUpdate, "User completed profile", &ip, &ua)

	// Redirect to dashboard
	redirectURL := getDashboardURLForRole(user)
	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
