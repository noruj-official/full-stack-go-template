// Package middleware provides HTTP middleware functions.
package middleware

import (
	"context"
	"net/http"

	"github.com/shaik-noor/full-stack-go-template/internal/domain"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
)

// Context keys for request context values.
type contextKey string

const (
	// UserContextKey is the key for storing the authenticated user in context.
	UserContextKey contextKey = "user"

	// SessionIDContextKey is the key for storing the session ID in context.
	SessionIDContextKey contextKey = "sessionID"
)

// SessionCookieName is the name of the session cookie.
const SessionCookieName = "session_id"

// Auth is middleware that validates the session and loads the user into context.
// It does not block access - use RequireAuth for protected routes.
type Auth struct {
	authService service.AuthService
}

// NewAuth creates a new auth middleware.
func NewAuth(authService service.AuthService) *Auth {
	return &Auth{authService: authService}
}

// Handler returns the middleware handler function.
func (a *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate session and get user
		user, err := a.authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
			next.ServeHTTP(w, r)
			return
		}

		// Add user and session ID to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, SessionIDContextKey, cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware ensures the user is authenticated.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/signin")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Check user status
		if user.Status != domain.UserStatusActive {
			// Clear any session if present
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/signin")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/signin?error=account_suspended", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware ensures the user has the required role.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/signin")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if user.HasPermission(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				if r.Header.Get("HX-Request") == "true" {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`<div class="text-red-500">Access denied</div>`))
					return
				}
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves the authenticated user from the request context.
func GetUserFromContext(ctx context.Context) *domain.User {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	if !ok {
		return nil
	}
	return user
}

// GetSessionIDFromContext retrieves the session ID from the request context.
func GetSessionIDFromContext(ctx context.Context) string {
	sessionID, ok := ctx.Value(SessionIDContextKey).(string)
	if !ok {
		return ""
	}
	return sessionID
}
