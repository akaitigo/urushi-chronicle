package handler

import (
	"net/http"
	"strings"

	"github.com/akaitigo/urushi-chronicle/internal/template"
)

// TemplateHandler handles HTTP requests for process step templates.
type TemplateHandler struct{}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

// ServeHTTP routes template requests.
// Paths:
//   - GET /api/v1/templates              — list all templates
//   - GET /api/v1/templates/{name}       — get a specific template
func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	const prefix = "/api/v1/templates"
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	rest = strings.TrimPrefix(rest, "/")

	if rest == "" {
		// List all templates
		writeJSON(w, http.StatusOK, template.AllWorkflows())
		return
	}

	// Get specific template
	wf := template.GetWorkflow(rest)
	writeJSON(w, http.StatusOK, wf)
}
