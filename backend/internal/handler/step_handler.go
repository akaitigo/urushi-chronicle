package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/akaitigo/urushi-chronicle/internal/storage"
	"github.com/google/uuid"
)

// StepHandler handles HTTP requests for process step CRUD operations.
type StepHandler struct {
	stepRepo repository.StepRepository
	workRepo repository.WorkRepository
	uploader storage.ImageUploader
}

// NewStepHandler creates a new StepHandler with the given dependencies.
func NewStepHandler(stepRepo repository.StepRepository, workRepo repository.WorkRepository, uploader storage.ImageUploader) *StepHandler {
	return &StepHandler{
		stepRepo: stepRepo,
		workRepo: workRepo,
		uploader: uploader,
	}
}

// createStepRequest is the JSON body for creating a process step.
type createStepRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	StepOrder     int      `json:"step_order"`
	Category      string   `json:"category"`
	MaterialsUsed []string `json:"materials_used"`
	Notes         string   `json:"notes"`
}

// updateStepRequest is the JSON body for updating a process step.
type updateStepRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	StepOrder     *int     `json:"step_order"`
	Category      *string  `json:"category"`
	MaterialsUsed []string `json:"materials_used"`
	Notes         *string  `json:"notes"`
	Completed     *bool    `json:"completed"`
}

// uploadRequest is the JSON body for requesting a presigned upload URL.
type uploadRequest struct {
	ContentType string `json:"content_type"`
}

// ServeHTTP routes requests to the appropriate handler method based on the URL path.
// Expected paths:
//   - POST   /api/v1/works/{workID}/steps
//   - GET    /api/v1/works/{workID}/steps
//   - GET    /api/v1/works/{workID}/steps/{stepID}
//   - PUT    /api/v1/works/{workID}/steps/{stepID}
//   - DELETE /api/v1/works/{workID}/steps/{stepID}
//   - POST   /api/v1/works/{workID}/steps/{stepID}/upload
func (h *StepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	workID, stepID, trailing, ok := h.parsePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid URL path")
		return
	}

	// Verify work exists
	if _, err := h.workRepo.FindByID(workID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Route based on presence of stepID and trailing path
	switch {
	case stepID == uuid.Nil && trailing == "":
		// /api/v1/works/{workID}/steps
		switch r.Method {
		case http.MethodPost:
			h.createStep(w, r, workID)
		case http.MethodGet:
			h.listSteps(w, workID)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	case stepID != uuid.Nil && trailing == "upload":
		// /api/v1/works/{workID}/steps/{stepID}/upload
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.requestUploadURL(w, r, workID, stepID)
	case stepID != uuid.Nil && trailing == "":
		// /api/v1/works/{workID}/steps/{stepID}
		switch r.Method {
		case http.MethodGet:
			h.getStep(w, workID, stepID)
		case http.MethodPut:
			h.updateStep(w, r, workID, stepID)
		case http.MethodDelete:
			h.deleteStep(w, workID, stepID)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// parsePath extracts workID, stepID, and trailing segment from the URL path.
// Path format: /api/v1/works/{workID}/steps[/{stepID}[/{trailing}]]
func (h *StepHandler) parsePath(path string) (workID, stepID uuid.UUID, trailing string, ok bool) {
	// Remove the prefix
	const prefix = "/api/v1/works/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, uuid.Nil, "", false
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")

	// Need at least: {workID}/steps
	if len(parts) < 2 || parts[1] != "steps" {
		return uuid.Nil, uuid.Nil, "", false
	}

	workID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, uuid.Nil, "", false
	}

	// /api/v1/works/{workID}/steps
	if len(parts) == 2 {
		return workID, uuid.Nil, "", true
	}

	// /api/v1/works/{workID}/steps/{stepID}
	stepID, err = uuid.Parse(parts[2])
	if err != nil {
		return uuid.Nil, uuid.Nil, "", false
	}

	if len(parts) == 3 {
		return workID, stepID, "", true
	}

	// /api/v1/works/{workID}/steps/{stepID}/{trailing}
	if len(parts) == 4 {
		return workID, stepID, parts[3], true
	}

	return uuid.Nil, uuid.Nil, "", false
}

func (h *StepHandler) createStep(w http.ResponseWriter, r *http.Request, workID uuid.UUID) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	now := time.Now().UTC()
	step := &domain.ProcessStep{
		ID:            uuid.New(),
		WorkID:        workID,
		Name:          req.Name,
		Description:   req.Description,
		StepOrder:     req.StepOrder,
		Category:      domain.StepCategory(req.Category),
		MaterialsUsed: req.MaterialsUsed,
		Notes:         req.Notes,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := step.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.stepRepo.Create(step); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create step")
		return
	}

	writeJSON(w, http.StatusCreated, step)
}

func (h *StepHandler) listSteps(w http.ResponseWriter, workID uuid.UUID) {
	steps, err := h.stepRepo.FindByWorkID(workID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list steps")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": steps,
		"total": len(steps),
	})
}

func (h *StepHandler) getStep(w http.ResponseWriter, workID, stepID uuid.UUID) {
	step, err := h.stepRepo.FindByID(workID, stepID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "step not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get step")
		return
	}

	writeJSON(w, http.StatusOK, step)
}

func (h *StepHandler) updateStep(w http.ResponseWriter, r *http.Request, workID, stepID uuid.UUID) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	existing, err := h.stepRepo.FindByID(workID, stepID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "step not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get step")
		return
	}

	var req updateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.StepOrder != nil {
		existing.StepOrder = *req.StepOrder
	}
	if req.Category != nil {
		existing.Category = domain.StepCategory(*req.Category)
	}
	if req.MaterialsUsed != nil {
		existing.MaterialsUsed = req.MaterialsUsed
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.Completed != nil && *req.Completed {
		now := time.Now().UTC()
		existing.CompletedAt = &now
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.stepRepo.Update(existing); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update step")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (h *StepHandler) deleteStep(w http.ResponseWriter, workID, stepID uuid.UUID) {
	if err := h.stepRepo.Delete(workID, stepID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "step not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete step")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StepHandler) requestUploadURL(w http.ResponseWriter, r *http.Request, workID, stepID uuid.UUID) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Verify step exists
	if _, err := h.stepRepo.FindByID(workID, stepID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "step not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.ContentType == "" {
		writeValidationErrors(w, []string{"content_type is required"})
		return
	}

	presigned, err := h.uploader.GenerateUploadURL(workID, stepID, req.ContentType)
	if err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, presigned)
}
