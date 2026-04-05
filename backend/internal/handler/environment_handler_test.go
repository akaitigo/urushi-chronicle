package handler_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/handler"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

// noopNotifier is a no-op implementation of alert.Notifier for tests.
type noopNotifier struct{}

func (n *noopNotifier) Notify(_ domain.EnvironmentReading, _ domain.AlertThreshold) error {
	return nil
}

func newTestEnvironmentHandler() (*handler.EnvironmentHandler, *repository.MemoryEnvironmentRepository, *repository.MemoryAlertThresholdRepository) {
	envRepo := repository.NewMemoryEnvironmentRepository()
	thresholdRepo := repository.NewMemoryAlertThresholdRepository()
	logger := log.New(os.Stderr, "[test] ", 0)
	notifier := &noopNotifier{}
	monitorSvc := monitor.NewService(envRepo, thresholdRepo, notifier, logger)
	h := handler.NewEnvironmentHandler(envRepo, thresholdRepo, monitorSvc)
	return h, envRepo, thresholdRepo
}

func TestEnvironmentHandler_IngestReading_Success(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":75.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp domain.EnvironmentReading
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.SensorID != "esp32-001" {
		t.Errorf("expected sensor_id esp32-001, got %s", resp.SensorID)
	}
}

func TestEnvironmentHandler_IngestReading_MissingFields(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"","location":"urushi_buro"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEnvironmentHandler_IngestReading_InvalidJSON(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEnvironmentHandler_IngestReading_OutOfRangeTemperature(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"esp32-001","location":"urushi_buro","temperature":101.0,"humidity":75.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEnvironmentHandler_QueryReadings(t *testing.T) {
	h, envRepo, _ := newTestEnvironmentHandler()

	// Seed data
	reading := &domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}
	if err := envRepo.Store(reading); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/readings?sensor_id=esp32-001", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestEnvironmentHandler_QueryReadings_MissingSensorID(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/readings", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEnvironmentHandler_CreateThreshold_Success(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"esp32-001","temperature_min":20.0,"temperature_max":30.0,"humidity_min":70.0,"humidity_max":85.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/thresholds", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp domain.AlertThreshold
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.SensorID != "esp32-001" {
		t.Errorf("expected sensor_id esp32-001, got %s", resp.SensorID)
	}
	if !resp.Enabled {
		t.Error("expected enabled to be true by default")
	}
}

func TestEnvironmentHandler_CreateThreshold_ValidationError(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	// temperature_min >= temperature_max
	body := `{"sensor_id":"esp32-001","temperature_min":30.0,"temperature_max":20.0,"humidity_min":70.0,"humidity_max":85.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/thresholds", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEnvironmentHandler_GetThreshold(t *testing.T) {
	h, _, thresholdRepo := newTestEnvironmentHandler()

	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	thresholdRepo.Seed(threshold)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/thresholds/"+threshold.ID.String(), nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestEnvironmentHandler_GetThreshold_NotFound(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/thresholds/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestEnvironmentHandler_DeleteThreshold(t *testing.T) {
	h, _, thresholdRepo := newTestEnvironmentHandler()

	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	thresholdRepo.Seed(threshold)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/environment/thresholds/"+threshold.ID.String(), nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEnvironmentHandler_ListThresholds(t *testing.T) {
	h, _, thresholdRepo := newTestEnvironmentHandler()

	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	thresholdRepo.Seed(threshold)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/thresholds", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestEnvironmentHandler_MethodNotAllowed(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"PUT readings", http.MethodPut, "/api/v1/environment/readings"},
		{"DELETE readings", http.MethodDelete, "/api/v1/environment/readings"},
		{"PATCH thresholds", http.MethodPatch, "/api/v1/environment/thresholds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", w.Code)
			}
		})
	}
}

func TestEnvironmentHandler_NotFound(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environment/unknown", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestEnvironmentHandler_IngestReading_WithTimestamp(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":75.0,"timestamp":"2026-03-29T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEnvironmentHandler_IngestReading_InvalidTimestamp(t *testing.T) {
	h, _, _ := newTestEnvironmentHandler()

	body := `{"sensor_id":"esp32-001","location":"urushi_buro","temperature":25.0,"humidity":75.0,"timestamp":"not-a-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/environment/readings", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
