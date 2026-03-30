package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

// EnvironmentHandler handles HTTP requests for environment monitoring endpoints.
type EnvironmentHandler struct {
	envRepo       repository.EnvironmentRepository
	thresholdRepo repository.AlertThresholdRepository
	monitorSvc    *monitor.Service
}

// NewEnvironmentHandler creates a new EnvironmentHandler with the given dependencies.
func NewEnvironmentHandler(
	envRepo repository.EnvironmentRepository,
	thresholdRepo repository.AlertThresholdRepository,
	monitorSvc *monitor.Service,
) *EnvironmentHandler {
	return &EnvironmentHandler{
		envRepo:       envRepo,
		thresholdRepo: thresholdRepo,
		monitorSvc:    monitorSvc,
	}
}

// ingestReadingRequest is the JSON body for manually ingesting a sensor reading via REST API.
type ingestReadingRequest struct {
	SensorID    string   `json:"sensor_id"`
	Location    string   `json:"location"`
	Temperature *float64 `json:"temperature"`
	Humidity    *float64 `json:"humidity"`
	Timestamp   string   `json:"timestamp,omitempty"`
}

// createThresholdRequest is the JSON body for creating an alert threshold.
type createThresholdRequest struct {
	SensorID       string  `json:"sensor_id"`
	TemperatureMin float64 `json:"temperature_min"`
	TemperatureMax float64 `json:"temperature_max"`
	HumidityMin    float64 `json:"humidity_min"`
	HumidityMax    float64 `json:"humidity_max"`
	Enabled        *bool   `json:"enabled"`
}

// ServeHTTP routes requests to the appropriate handler method.
// Expected paths:
//   - POST /api/v1/environment/readings        — ingest a reading via REST
//   - GET  /api/v1/environment/readings        — query readings by sensor_id
//   - POST /api/v1/environment/thresholds      — create an alert threshold
//   - GET  /api/v1/environment/thresholds      — list thresholds
//   - GET  /api/v1/environment/thresholds/{id} — get a threshold
//   - PUT  /api/v1/environment/thresholds/{id} — update a threshold
//   - DELETE /api/v1/environment/thresholds/{id} — delete a threshold
func (h *EnvironmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/environment/")

	switch {
	case path == "readings":
		h.handleReadings(w, r)
	case path == "thresholds":
		h.handleThresholds(w, r)
	case strings.HasPrefix(path, "thresholds/"):
		idStr := strings.TrimPrefix(path, "thresholds/")
		h.handleThresholdByID(w, r, idStr)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *EnvironmentHandler) handleReadings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.ingestReading(w, r)
	case http.MethodGet:
		h.queryReadings(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *EnvironmentHandler) handleThresholds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createThreshold(w, r)
	case http.MethodGet:
		h.listThresholds(w)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *EnvironmentHandler) handleThresholdByID(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid threshold ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getThreshold(w, id)
	case http.MethodPut:
		h.updateThreshold(w, r, id)
	case http.MethodDelete:
		h.deleteThreshold(w, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *EnvironmentHandler) ingestReading(w http.ResponseWriter, r *http.Request) {
	var req ingestReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var validationErrors []string
	if req.SensorID == "" {
		validationErrors = append(validationErrors, "sensor_id is required")
	}
	if req.Location == "" {
		validationErrors = append(validationErrors, "location is required")
	}
	if req.Temperature == nil {
		validationErrors = append(validationErrors, "temperature is required")
	}
	if req.Humidity == nil {
		validationErrors = append(validationErrors, "humidity is required")
	}
	if len(validationErrors) > 0 {
		writeValidationErrors(w, validationErrors)
		return
	}

	t := time.Now().UTC()
	if req.Timestamp != "" {
		parsed, parseErr := time.Parse(time.RFC3339, req.Timestamp)
		if parseErr != nil {
			writeValidationErrors(w, []string{"timestamp must be in RFC3339 format"})
			return
		}
		t = parsed.UTC()
	}

	reading := domain.EnvironmentReading{
		Time:        t,
		SensorID:    req.SensorID,
		Location:    req.Location,
		Temperature: *req.Temperature,
		Humidity:    *req.Humidity,
	}

	if err := h.monitorSvc.ProcessReading(reading); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, reading)
}

func (h *EnvironmentHandler) queryReadings(w http.ResponseWriter, r *http.Request) {
	sensorID := r.URL.Query().Get("sensor_id")
	if sensorID == "" {
		writeValidationErrors(w, []string{"sensor_id query parameter is required"})
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 1000 {
			writeValidationErrors(w, []string{"limit must be an integer between 1 and 1000"})
			return
		}
		limit = parsed
	}

	readings, err := h.envRepo.FindBySensorID(sensorID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query readings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": readings,
		"total": len(readings),
	})
}

func (h *EnvironmentHandler) createThreshold(w http.ResponseWriter, r *http.Request) {
	var req createThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now().UTC()
	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       req.SensorID,
		TemperatureMin: req.TemperatureMin,
		TemperatureMax: req.TemperatureMax,
		HumidityMin:    req.HumidityMin,
		HumidityMax:    req.HumidityMax,
		Enabled:        enabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := threshold.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.thresholdRepo.Create(threshold); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create threshold")
		return
	}

	writeJSON(w, http.StatusCreated, threshold)
}

func (h *EnvironmentHandler) listThresholds(w http.ResponseWriter) {
	thresholds, err := h.thresholdRepo.FindAllEnabled()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list thresholds")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": thresholds,
		"total": len(thresholds),
	})
}

func (h *EnvironmentHandler) getThreshold(w http.ResponseWriter, id uuid.UUID) {
	threshold, err := h.thresholdRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "threshold not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get threshold")
		return
	}
	writeJSON(w, http.StatusOK, threshold)
}

func (h *EnvironmentHandler) updateThreshold(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	existing, err := h.thresholdRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "threshold not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get threshold")
		return
	}

	var req createThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.SensorID != "" {
		existing.SensorID = req.SensorID
	}
	if req.TemperatureMin != 0 || req.TemperatureMax != 0 {
		existing.TemperatureMin = req.TemperatureMin
		existing.TemperatureMax = req.TemperatureMax
	}
	if req.HumidityMin != 0 || req.HumidityMax != 0 {
		existing.HumidityMin = req.HumidityMin
		existing.HumidityMax = req.HumidityMax
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		writeValidationErrors(w, []string{err.Error()})
		return
	}

	if err := h.thresholdRepo.Update(existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update threshold")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (h *EnvironmentHandler) deleteThreshold(w http.ResponseWriter, id uuid.UUID) {
	if err := h.thresholdRepo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "threshold not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete threshold")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
