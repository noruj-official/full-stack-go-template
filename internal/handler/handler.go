// Package handler provides HTTP request handlers.
package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/a-h/templ"
	"github.com/shaik-noor/full-stack-go-template/internal/middleware"
)

// Handler is the base handler with shared utilities.
type Handler struct {
	templates     map[string]*template.Template
	templatesOnce sync.Once
	templatesDir  string
	appName       string
	appLogo       string
}

// NewHandler creates a new base handler.
func NewHandler(templatesDir, appName, appLogo string) *Handler {
	return &Handler{
		templatesDir: templatesDir,
		templates:    make(map[string]*template.Template),
		appName:      appName,
		appLogo:      appLogo,
	}
}

// LoadTemplates parses all templates with the base layout.
func (h *Handler) LoadTemplates() error {
	var err error
	h.templatesOnce.Do(func() {
		baseLayout := filepath.Join(h.templatesDir, "layouts", "base.html")

		// Add template functions
		funcMap := template.FuncMap{
			"add":      func(a, b int) int { return a + b },
			"subtract": func(a, b int) int { return a - b },
			"sub":      func(a, b int) int { return a - b },
			"mul":      func(a, b int) int { return a * b },
			"div": func(a, b int) int {
				if b == 0 {
					return 0
				}
				return a / b
			},
			"slice": func(s string, start, end int) string {
				// rune-aware slicing to handle multibyte characters safely
				r := []rune(s)
				if start < 0 {
					start = 0
				}
				if end > len(r) {
					end = len(r)
				}
				if start > end {
					return ""
				}
				return string(r[start:end])
			},
		}

		// Parse component templates
		componentFiles, _ := filepath.Glob(filepath.Join(h.templatesDir, "components", "*.html"))

		// Parse page templates
		pageFiles, _ := filepath.Glob(filepath.Join(h.templatesDir, "pages", "*.html"))
		userPageFiles, _ := filepath.Glob(filepath.Join(h.templatesDir, "pages", "users", "*.html"))
		dashboardFiles, _ := filepath.Glob(filepath.Join(h.templatesDir, "pages", "dashboards", "*.html"))
		pageFiles = append(pageFiles, userPageFiles...)
		pageFiles = append(pageFiles, dashboardFiles...)

		// Parse partial templates
		partialFiles, _ := filepath.Glob(filepath.Join(h.templatesDir, "partials", "*.html"))

		allFiles := append(componentFiles, partialFiles...)

		for _, page := range pageFiles {
			files := append([]string{baseLayout, page}, allFiles...)
			name := filepath.Base(page)

			// Use template.New with Funcs to register custom functions
			tmpl, parseErr := template.New("base").Funcs(funcMap).ParseFiles(files...)
			if parseErr != nil {
				err = parseErr
				log.Printf("Error parsing template %s: %v", name, parseErr)
				continue
			}
			h.templates[name] = tmpl
		}

		// Parse partials independently for HTMX responses
		for _, partial := range partialFiles {
			name := filepath.Base(partial)
			tmpl, parseErr := template.New(name).Funcs(funcMap).ParseFiles(partial)
			if parseErr != nil {
				log.Printf("Error parsing partial %s: %v", name, parseErr)
				continue
			}
			h.templates["partial:"+name] = tmpl
		}
	})
	return err
}

// Render renders a template with the given data.
func (h *Handler) Render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := h.templates[name]
	if !ok {
		log.Printf("Template %s not found", name)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// RenderWithUser merges the authenticated user from context into the data and renders.
func (h *Handler) RenderWithUser(w http.ResponseWriter, r *http.Request, name string, data any) {
	if data == nil {
		data = map[string]any{}
	}
	if m, ok := data.(map[string]any); ok {
		m["User"] = middleware.GetUserFromContext(r.Context())
		m["AppName"] = h.appName
		m["AppLogo"] = h.appLogo
		m["AppName"] = h.appName
		m["AppLogo"] = h.appLogo
		// Inject Theme
		m["Theme"] = h.GetTheme(r)
		h.Render(w, name, m)
		return
	}
	// For non-map data (e.g., structs used in partials), render as-is
	h.Render(w, name, data)
}

// GetTheme extracts the theme from the request (cookie or client hint).
func (h *Handler) GetTheme(r *http.Request) string {
	theme := "light"
	if c, err := r.Cookie("theme"); err == nil && c.Value != "" {
		if c.Value == "dark" {
			theme = "dark"
		} else if c.Value == "light" {
			theme = "light"
		}
	} else {
		// Optional: use client hint if sent
		if v := r.Header.Get("Sec-CH-Prefers-Color-Scheme"); v == "dark" {
			theme = "dark"
		}
	}
	return theme
}

// RenderPartial renders a partial template (for HTMX responses).
func (h *Handler) RenderPartial(w http.ResponseWriter, name string, data any) {
	tmpl, ok := h.templates["partial:"+name]
	if !ok {
		// Attempt on-demand parse of the partial if it was added after initial load
		partialPath := filepath.Join(h.templatesDir, "partials", name)
		funcMap := template.FuncMap{
			"add":      func(a, b int) int { return a + b },
			"subtract": func(a, b int) int { return a - b },
			"sub":      func(a, b int) int { return a - b },
			"mul":      func(a, b int) int { return a * b },
			"div": func(a, b int) int {
				if b == 0 {
					return 0
				}
				return a / b
			},
			"slice": func(s string, start, end int) string {
				// rune-aware slicing to handle multibyte characters safely
				r := []rune(s)
				if start < 0 {
					start = 0
				}
				if end > len(r) {
					end = len(r)
				}
				if start > end {
					return ""
				}
				return string(r[start:end])
			},
		}
		parsed, err := template.New(name).Funcs(funcMap).ParseFiles(partialPath)
		if err != nil {
			log.Printf("Partial template %s not found", name)
			http.Error(w, "Partial template not found", http.StatusInternalServerError)
			return
		}
		h.templates["partial:"+name] = parsed
		tmpl = parsed
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering partial %s: %v", name, err)
		http.Error(w, "Error rendering partial", http.StatusInternalServerError)
	}
}

// RenderPartialWithUser merges the authenticated user into data and renders a partial.
func (h *Handler) RenderPartialWithUser(w http.ResponseWriter, r *http.Request, name string, data any) {
	if data == nil {
		data = map[string]any{}
	}
	if m, ok := data.(map[string]any); ok {
		m["User"] = middleware.GetUserFromContext(r.Context())
		h.RenderPartial(w, name, m)
		return
	}
	// For non-map data (e.g., structs used in row partials), render as-is
	h.RenderPartial(w, name, data)
}

// JSON sends a JSON response.
func (h *Handler) JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// Error sends an error response, automatically detecting HTMX requests.
func (h *Handler) Error(w http.ResponseWriter, r *http.Request, status int, message string) {
	if isHTMXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		w.Write([]byte(`<div class="text-red-500">` + message + `</div>`))
		return
	}

	http.Error(w, message, status)
}

// isHTMXRequest checks if the request is from HTMX.
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func isHTMXBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

// RenderTempl renders a templ component.
func (h *Handler) RenderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering templ component: %v", err)
		h.Error(w, r, http.StatusInternalServerError, "Error rendering template")
	}
}
