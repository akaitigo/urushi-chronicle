package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func validAlertThreshold() *domain.AlertThreshold {
	return &domain.AlertThreshold{
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
}

func TestMemoryAlertThresholdRepository_CreateAndFindByID(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()
	threshold := validAlertThreshold()

	if err := repo.Create(context.Background(), threshold); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found, err := repo.FindByID(context.Background(), threshold.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.SensorID != threshold.SensorID {
		t.Errorf("expected sensor_id %s, got %s", threshold.SensorID, found.SensorID)
	}
}

func TestMemoryAlertThresholdRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryAlertThresholdRepository_FindBySensorID(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()

	t1 := validAlertThreshold()
	t1.SensorID = "esp32-001"
	t1.Enabled = true

	t2 := validAlertThreshold()
	t2.SensorID = "esp32-001"
	t2.Enabled = false // disabled — should be excluded

	t3 := validAlertThreshold()
	t3.SensorID = "esp32-002"
	t3.Enabled = true

	for _, th := range []*domain.AlertThreshold{t1, t2, t3} {
		if err := repo.Create(context.Background(), th); err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	results, err := repo.FindBySensorID(context.Background(), "esp32-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 enabled threshold for esp32-001, got %d", len(results))
	}
}

func TestMemoryAlertThresholdRepository_FindAllEnabled(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()

	enabled := validAlertThreshold()
	enabled.Enabled = true
	disabled := validAlertThreshold()
	disabled.Enabled = false

	for _, th := range []*domain.AlertThreshold{enabled, disabled} {
		if err := repo.Create(context.Background(), th); err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	results, err := repo.FindAllEnabled(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 enabled threshold, got %d", len(results))
	}
}

func TestMemoryAlertThresholdRepository_Update(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()
	threshold := validAlertThreshold()
	if err := repo.Create(context.Background(), threshold); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	threshold.TemperatureMax = 35.0
	if err := repo.Update(context.Background(), threshold); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	found, err := repo.FindByID(context.Background(), threshold.ID)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if found.TemperatureMax != 35.0 {
		t.Errorf("expected temperature_max 35.0, got %f", found.TemperatureMax)
	}
}

func TestMemoryAlertThresholdRepository_Update_NotFound(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()
	threshold := validAlertThreshold()

	err := repo.Update(context.Background(), threshold)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryAlertThresholdRepository_Delete(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()
	threshold := validAlertThreshold()
	if err := repo.Create(context.Background(), threshold); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := repo.Delete(context.Background(), threshold.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err := repo.FindByID(context.Background(), threshold.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryAlertThresholdRepository_Delete_NotFound(t *testing.T) {
	repo := repository.NewMemoryAlertThresholdRepository()

	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
