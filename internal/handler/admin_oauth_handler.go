package handler

import (
	"net/http"
	"strings"

	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/repository"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	adminPage "github.com/noruj-official/full-stack-go-template/web/templ/pages/admin"
)

type AdminOAuthHandler struct {
	*Handler
	oauthRepo    repository.OAuthRepository
	auditService service.AuditService
	appURL       string
}

func NewAdminOAuthHandler(base *Handler, oauthRepo repository.OAuthRepository, auditService service.AuditService, appURL string) *AdminOAuthHandler {
	return &AdminOAuthHandler{
		Handler:      base,
		oauthRepo:    oauthRepo,
		auditService: auditService,
		appURL:       appURL,
	}
}

// List renders the OAuth settings page.
func (h *AdminOAuthHandler) List(w http.ResponseWriter, r *http.Request) {
	providers, err := h.oauthRepo.ListProviders(r.Context())
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load oauth providers")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, adminPage.OAuthSettings(user, theme, themeEnabled, oauthEnabled, providers, h.appURL))
}

// Update handles the update of an OAuth provider.
func (h *AdminOAuthHandler) Update(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	if providerName == "" {
		http.Error(w, "Provider name required", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	scopesStr := r.FormValue("scopes")
	authURL := r.FormValue("auth_url")
	tokenURL := r.FormValue("token_url")
	userInfoURL := r.FormValue("user_info_url")
	enabled := r.FormValue("enabled") == "on"

	// Prevent disabling the last active OAuth provider if OAuth is the only auth method enabled
	if !enabled {
		// 1. Check if OAuth feature is enabled
		oauthFeatureEnabled, err := h.featureService.IsEnabled(r.Context(), domain.FeatureOAuth)
		if err == nil && oauthFeatureEnabled {
			// 2. Check if other auth methods are disabled
			emailAuthEnabled, _ := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailAuth)
			passwordAuthEnabled, _ := h.featureService.IsEnabled(r.Context(), domain.FeatureEmailPasswordAuth)

			if !emailAuthEnabled && !passwordAuthEnabled {
				// 3. OAuth is the only method. Check if this is the last provider.
				providers, err := h.oauthRepo.ListProviders(r.Context())
				if err == nil {
					activeCount := 0
					for _, p := range providers {
						if p.Enabled && string(p.Provider) != providerName {
							activeCount++
						}
					}
					if activeCount == 0 {
						// This is the last one!
						// Trigger error toast and return current state (re-render card as is, effectively reverting)
						existing, err := h.oauthRepo.GetProvider(r.Context(), domain.OAuthProviderType(providerName))
						if err != nil {
							http.Error(w, "Provider not found", http.StatusNotFound)
							return
						}
						w.Header().Set("HX-Trigger", `{"error-toast": "At least one OAuth provider must be enabled when OAuth is the only authentication method"}`)
						h.RenderTempl(w, r, adminPage.OAuthProviderCard(existing, h.appURL))
						return
					}
				}
			}
		}
	}

	// Fetch existing to keep urls and secret if empty?
	existing, err := h.oauthRepo.GetProvider(r.Context(), domain.OAuthProviderType(providerName))
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Update fields
	existing.ClientID = clientID
	if clientSecret != "" {
		existing.ClientSecret = clientSecret
	}
	existing.Enabled = enabled
	existing.AuthURL = authURL
	existing.TokenURL = tokenURL
	existing.UserInfoURL = userInfoURL

	// Split scopes
	var scopes []string
	if scopesStr != "" {
		parts := strings.Split(scopesStr, ",")
		for _, p := range parts {
			scopes = append(scopes, strings.TrimSpace(p))
		}
	}
	existing.Scopes = scopes

	if err := h.oauthRepo.UpdateProvider(r.Context(), existing); err != nil {
		http.Error(w, "Failed to update provider", http.StatusInternalServerError)
		return
	}

	// Log audit
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, "update_oauth_provider", "oauth_provider", nil, nil, map[string]interface{}{
			"provider": providerName,
			"enabled":  enabled,
		}, &ip)
	}

	// Render the updated card
	h.RenderTempl(w, r, adminPage.OAuthProviderCard(existing, h.appURL))
}
