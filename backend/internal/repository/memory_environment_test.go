package repository_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
)

func TestMemoryEnvironmentRepository_Store(t *testing.T) {
	repo := repository.NewMemoryEnvironmentRepository()
	reading := &domain.EnvironmentReading{
		Time:        time.Now().UTC(),
		SensorID:    "esp32-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}

	if err := repo.Store(reading); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	readings, err := repo.FindBySensorID("esp32-001", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("expected 1 reading, got %d", len(readings))
	}
	if readings[0].Temperature != 25.0 {
		t.Errorf("expected temperature 25.0, got %f", readings[0].Temperature)
	}
}

func TestMemoryEnvironmentRepository_FindBySensorID_NoResults(t *testing.T) {
	repo := repository.NewMemoryEnvironmentRepository()

	readings, err := repo.FindBySensorID("nonexistent", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(readings) != 0 {
		t.Errorf("expected empty slice, got %d readings", len(readings))
	}
}

func TestMemoryEnvironmentRepository_FindBySensorID_Limit(t *testing.T) {
	repo := repository.NewMemoryEnvironmentRepository()

	base := time.Now().UTC()
	for i := 0; i < 5; i++ {
		reading := &domain.EnvironmentReading{
			Time:        base.Add(time.Duration(i) * time.Minute),
			SensorID:    "esp32-001",
			Location:    "urushi_buro",
			Temperature: 20.0 + float64(i),
			Humidity:    70.0,
		}
		if err := repo.Store(reading); err != nil {
			t.Fatalf("store failed: %v", err)
		}
	}

	readings, err := repo.FindBySensorID("esp32-001", 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(readings) != 3 {
		t.Fatalf("expected 3 readings, got %d", len(readings))
	}
	// Should be ordered by time descending (most recent first)
	if readings[0].Temperature != 24.0 {
		t.Errorf("expected most recent reading first (temperature 24.0), got %f", readings[0].Temperature)
	}
}

func TestMemoryEnvironmentRepository_FindBySensorID_FiltersBySensor(t *testing.T) {
	repo := repository.NewMemoryEnvironmentRepository()

	for _, sensorID := range []string{"esp32-001", "esp32-002", "esp32-001"} {
		reading := &domain.EnvironmentReading{
			Time:        time.Now().UTC(),
			SensorID:    sensorID,
			Location:    "urushi_buro",
			Temperature: 25.0,
			Humidity:    75.0,
		}
		if err := repo.Store(reading); err != nil {
			t.Fatalf("store failed: %v", err)
		}
	}

	readings, err := repo.FindBySensorID("esp32-001", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(readings) != 2 {
		t.Errorf("expected 2 readings for esp32-001, got %d", len(readings))
	}
}
