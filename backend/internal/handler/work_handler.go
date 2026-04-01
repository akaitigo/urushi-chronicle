package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
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

// createWorkRequest is the JSON body for POST /api/v1/works.
type createWorkRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Technique   string `json:"technique"`
	Material    string `json:"material"`
}

// updateWorkRequest is the JSON body for PUT /api/v1/works/{id}.
type updateWorkRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Technique   string  `json:"technique"`
	Material    string  `json:"material"`
	Status      string  `json:"status"`
	CompletedAt *string `json:"completed_at"`
}

// ServeHTTP routes requests to the appropriate handler method.
// Expected paths:
//   - GET    /api/v1/works          — list all works
//   - POST   /api/v1/works          — create a work
//   - GET    /api/v1/works/{id}     — get a single work by ID
//   - PUT    /api/v1/works/{id}     — update a work
//   - DELETE /api/v1/works/{id}     — delete a work
func (h *WorkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const prefix = "/api/v1/works"
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimPrefix(path, "/")

	// Collection endpoints (no ID segment)
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.listWorks(w)
		case http.MethodPost:
			h.createWork(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
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

	// Single-resource endpoints
	switch r.Method {
	case http.MethodGet:
		h.getWork(w, id)
	case http.MethodPut:
		h.updateWork(w, r, id)
	case http.MethodDelete:
		h.deleteWork(w, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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

func (h *WorkHandler) createWork(w http.ResponseWriter, r *http.Request) {
	var req createWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	now := time.Now().UTC()
	work := &domain.Work{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Technique:   domain.Technique(req.Technique),
		Material:    req.Material,
		Status:      domain.WorkStatusInProgress,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := work.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.workRepo.Create(work); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create work")
		return
	}

	writeJSON(w, http.StatusCreated, work)
}

func (h *WorkHandler) updateWork(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	existing, err := h.workRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get work")
		return
	}

	var req updateWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Apply partial updates: only overwrite fields that are non-zero
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Technique != "" {
		existing.Technique = domain.Technique(req.Technique)
	}
	if req.Material != "" {
		existing.Material = req.Material
	}
	if req.Status != "" {
		existing.Status = domain.WorkStatus(req.Status)
	}
	if req.CompletedAt != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.CompletedAt)
		if parseErr != nil {
			writeValidationErrors(w, []string{"completed_at must be RFC3339 format"})
			return
		}
		existing.CompletedAt = &t
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.workRepo.Update(existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update work")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (h *WorkHandler) deleteWork(w http.ResponseWriter, id uuid.UUID) {
	if err := h.workRepo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete work")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
