package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

// WorkHandler handles HTTP requests for work (lacquerware piece) endpoints.
type WorkHandler struct {
	workRepo repository.WorkRepository
}

// NewWorkHandler creates a new WorkHandler with the given repository.
func NewWorkHandler(workRepo repository.WorkRepository) *WorkHandler {
	return &WorkHandler{workRepo: workRepo}
}

// ServeHTTP routes requests to the appropriate handler method.
// Expected paths:
//   - GET /api/v1/works          — list all works
//   - GET /api/v1/works/{id}     — get a single work by ID
func (h *WorkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	const prefix = "/api/v1/works"
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		h.listWorks(w)
		return
	}

	// Reject paths with extra segments (e.g., /api/v1/works/{id}/something)
	if strings.Contains(path, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	id, err := uuid.Parse(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work ID")
		return
	}
	h.getWork(w, id)
}

func (h *WorkHandler) listWorks(w http.ResponseWriter) {
	works, err := h.workRepo.FindAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list works")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": works,
		"total": len(works),
	})
}

func (h *WorkHandler) getWork(w http.ResponseWriter, id uuid.UUID) {
	work, err := h.workRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get work")
		return
	}

	writeJSON(w, http.StatusOK, work)
}
