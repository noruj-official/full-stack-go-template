package handler

import (
	"net/http"

	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	featuresPage "github.com/noruj-official/full-stack-go-template/web/templ/pages/features"
)

// FeatureHandler handles feature flag related requests.
type FeatureHandler struct {
	*Handler
	featureService service.FeatureService
	auditService   service.AuditService
}

// NewFeatureHandler creates a new feature handler.
func NewFeatureHandler(base *Handler, featureService service.FeatureService, auditService service.AuditService) *FeatureHandler {
	return &FeatureHandler{
		Handler:        base,
		featureService: featureService,
		auditService:   auditService,
	}
}

// List renders the feature flags list page.
func (h *FeatureHandler) List(w http.ResponseWriter, r *http.Request) {
	features, err := h.featureService.GetAll(r.Context())
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load feature flags")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	showSidebar := true

	h.RenderTempl(w, r, featuresPage.List("Feature Flags", "Manage application feature flags", user, showSidebar, theme, themeEnabled, oauthEnabled, features))
}

// Toggle handles the HTMX toggle of a feature flag.
func (h *FeatureHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Feature name required", http.StatusBadRequest)
		return
	}

	// Get current state to flush logic? Or just check form/query?
	// For a toggle switch, usually we receive the new desired state or we flip current.
	// But let's assume the UI sends the desired state via form value or we just flip it.
	// Let's rely on a 'enabled' query param or form value.
	enabledStr := r.URL.Query().Get("enabled")
	enabled := enabledStr == "true"

	if err := h.featureService.Toggle(r.Context(), name, enabled); err != nil {
		if err == domain.ErrAtLeastOneAuthMethodRequired {
			// Fetch the current feature state (which should be unchanged)
			feature, getErr := h.featureService.Get(r.Context(), name)
			if getErr != nil {
				http.Error(w, "Failed to retrieve feature", http.StatusInternalServerError)
				return
			}

			// Trigger an error toast and re-render the row
			w.Header().Set("HX-Trigger", `{"error-toast": "At least one authentication method must be enabled"}`)
			h.RenderTempl(w, r, featuresPage.FeatureRow(feature))
			return
		}

		http.Error(w, "Failed to toggle feature", http.StatusInternalServerError)
		return
	}

	// Retrieve updated flag to render row
	feature, err := h.featureService.Get(r.Context(), name)
	if err != nil {
		// Log error but we already toggled successfully so...
		// Ideally we shouldn't fail here.
		// If we can't get it, we could just return OK, but the UI might break.
		http.Error(w, "Failed to retrieve updated feature", http.StatusInternalServerError)
		return
	}

	// Log audit
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, "feature_toggle", "feature_flag", nil, nil, map[string]interface{}{
			"name":    name,
			"enabled": enabled,
		}, &ip)
	}

	w.Header().Set("HX-Trigger", "featureToggled")
	// Render the updated row
	h.RenderTempl(w, r, featuresPage.FeatureRow(feature))
}
