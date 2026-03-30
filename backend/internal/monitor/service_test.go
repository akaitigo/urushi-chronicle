package monitor_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/monitor"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

// mockNotifier records alert notifications for testing.
type mockNotifier struct {
	alerts []notifiedAlert
}

type notifiedAlert struct {
	reading   domain.EnvironmentReading
	threshold domain.AlertThreshold
}

func (m *mockNotifier) Notify(reading domain.EnvironmentReading, threshold domain.AlertThreshold) error {
	m.alerts = append(m.alerts, notifiedAlert{reading: reading, threshold: threshold})
	return nil
}

func newTestService() (*monitor.Service, *repository.MemoryEnvironmentRepository, *repository.MemoryAlertThresholdRepository, *mockNotifier) {
	envRepo := repository.NewMemoryEnvironmentRepository()
	thresholdRepo := repository.NewMemoryAlertThresholdRepository()
	notifier := &mockNotifier{}
	logger := log.New(os.Stderr, "[test] ", 0)
	svc := monitor.NewService(envRepo, thresholdRepo, notifier, logger)
	return svc, envRepo, thresholdRepo, notifier
}

func TestService_ProcessReading_StoresReading(t *testing.T) {
	svc, envRepo, _, _ := newTestService()

	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stored, err := envRepo.FindBySensorID("esp32-001", 10)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected 1 stored reading, got %d", len(stored))
	}
	if stored[0].Temperature != 25.0 {
		t.Errorf("expected temperature 25.0, got %f", stored[0].Temperature)
	}
}

func TestService_ProcessReading_TriggersAlert(t *testing.T) {
	svc, _, thresholdRepo, notifier := newTestService()

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

	// Temperature exceeds maximum
	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 35.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifier.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(notifier.alerts))
	}
	if notifier.alerts[0].threshold.ID != threshold.ID {
		t.Errorf("expected threshold ID %s, got %s", threshold.ID, notifier.alerts[0].threshold.ID)
	}
}

func TestService_ProcessReading_NoAlertWhenWithinRange(t *testing.T) {
	svc, _, thresholdRepo, notifier := newTestService()

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

	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifier.alerts) != 0 {
		t.Errorf("expected no alerts when within range, got %d", len(notifier.alerts))
	}
}

func TestService_ProcessReading_NoAlertForDisabledThreshold(t *testing.T) {
	svc, _, thresholdRepo, notifier := newTestService()

	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-001",
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        false, // disabled
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	thresholdRepo.Seed(threshold)

	// Temperature exceeds maximum, but threshold is disabled
	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 35.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifier.alerts) != 0 {
		t.Errorf("expected no alerts for disabled threshold, got %d", len(notifier.alerts))
	}
}

func TestService_ProcessReading_NoAlertForDifferentSensor(t *testing.T) {
	svc, _, thresholdRepo, notifier := newTestService()

	threshold := &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       "esp32-002", // different sensor
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	thresholdRepo.Seed(threshold)

	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001", // different sensor
		Location:    "urushi_buro",
		Temperature: 35.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifier.alerts) != 0 {
		t.Errorf("expected no alerts for different sensor, got %d", len(notifier.alerts))
	}
}

func TestService_ProcessReading_InvalidReading(t *testing.T) {
	svc, _, _, _ := newTestService()

	// Empty sensor_id should fail validation
	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}

	if err := svc.ProcessReading(reading); err == nil {
		t.Error("expected error for invalid reading")
	}
}

func TestService_ProcessReading_HumidityAlert(t *testing.T) {
	svc, _, thresholdRepo, notifier := newTestService()

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

	// Humidity below minimum
	reading := domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    50.0, // below 70.0 minimum
	}

	if err := svc.ProcessReading(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifier.alerts) != 1 {
		t.Errorf("expected 1 alert for low humidity, got %d", len(notifier.alerts))
	}
}
