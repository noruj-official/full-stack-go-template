package middleware

import (
	"net/http"
)

// RequireVerifiedEmail middleware ensures the user's email is verified.
func RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			// Not authenticated, let standard auth middleware handle it or redirect
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if !user.EmailVerified {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/verify-email")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/verify-email", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
