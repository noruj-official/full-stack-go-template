package handler

import (
	"net/http"

	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
	"github.com/shaik-noor/full-stack-go-template/internal/service"
	featuresPage "github.com/shaik-noor/full-stack-go-template/web/templ/pages/features"
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
	showSidebar := true

	h.RenderTempl(w, r, featuresPage.List("Feature Flags", "Manage application feature flags", user, showSidebar, theme, themeEnabled, features))
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
		http.Error(w, "Failed to toggle feature", http.StatusInternalServerError)
		return
	}

	// Retrieve updated flag to render row
	// Retrieve updated flag to render row
	// feature, _ := h.featureService.IsEnabled(r.Context(), name)
	// We might need the full object to render the row, but IsEnabled only returns bool.
	// Let's refactor IsEnabled or just construct a temporary object or fetch via repository in service if needed.
	// But wait, we have GetAll. We can fetch specific one if we exposed it.
	// actually for a toggle, we usually just return the new switch HTML or row.
	// For simplicity, let's just re-render the list or row if we had a Get(id) in service.
	// IsEnabled does DB call. Let's add Get to service? No, IsEnabled is fine for logic.
	// Let's just assume success and return success message or updated switch.

	// Since we don't have Get(name) exposed in service interface publically (it is in repo),
	// let's just rely on the fact we know the new state.

	// Better yet, let's respond with the updated toggle switch component.
	// But for now, let's just redirect or return OK if HTMX.

	// Log audit
	if admin := middleware.GetUserFromContext(r.Context()); admin != nil {
		ip := getIPAddress(r)
		_ = h.auditService.LogAudit(r.Context(), admin.ID, "feature_toggle", "feature_flag", nil, nil, map[string]interface{}{
			"name":    name,
			"enabled": enabled,
		}, &ip)
	}

	w.Header().Set("HX-Trigger", "featureToggled")
	w.WriteHeader(http.StatusOK)
}
