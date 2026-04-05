package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPgAlertThresholdRepo(t *testing.T) *repository.PgAlertThresholdRepository {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("failed to create pool (DB not available): %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("failed to ping database (DB not available): %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	return repository.NewPgAlertThresholdRepository(pool)
}

func newPgTestThreshold(sensorID string) *domain.AlertThreshold {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &domain.AlertThreshold{
		ID:             uuid.New(),
		SensorID:       sensorID,
		TemperatureMin: 20.0,
		TemperatureMax: 30.0,
		HumidityMin:    70.0,
		HumidityMax:    85.0,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func TestPgAlertThresholdRepository_CRUD(t *testing.T) {
	repo := setupPgAlertThresholdRepo(t)

	threshold := newPgTestThreshold("test-alert-sensor")
	t.Cleanup(func() {
		_ = repo.Delete(threshold.ID)
	})

	// Create
	if err := repo.Create(threshold); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(threshold.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.SensorID != "test-alert-sensor" {
		t.Errorf("expected sensor_id 'test-alert-sensor', got %q", found.SensorID)
	}

	// FindBySensorID
	results, err := repo.FindBySensorID("test-alert-sensor")
	if err != nil {
		t.Fatalf("FindBySensorID failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least 1 threshold from FindBySensorID")
	}

	// FindAllEnabled
	all, err := repo.FindAllEnabled()
	if err != nil {
		t.Fatalf("FindAllEnabled failed: %v", err)
	}
	if len(all) == 0 {
		t.Error("expected at least 1 enabled threshold")
	}

	// Update
	found.TemperatureMax = 35.0
	found.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(found); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	updated, err := repo.FindByID(threshold.ID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if updated.TemperatureMax != 35.0 {
		t.Errorf("expected TemperatureMax 35.0, got %f", updated.TemperatureMax)
	}

	// Delete
	if err := repo.Delete(threshold.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.FindByID(threshold.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPgAlertThresholdRepository_FindAndUpdate(t *testing.T) {
	repo := setupPgAlertThresholdRepo(t)

	threshold := newPgTestThreshold("test-fau-sensor")
	t.Cleanup(func() {
		_ = repo.Delete(threshold.ID)
	})

	if err := repo.Create(threshold); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	updated, err := repo.FindAndUpdate(threshold.ID, func(existing *domain.AlertThreshold) (*domain.AlertThreshold, error) {
		existing.HumidityMax = 90.0
		existing.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		return existing, nil
	})
	if err != nil {
		t.Fatalf("FindAndUpdate failed: %v", err)
	}
	if updated.HumidityMax != 90.0 {
		t.Errorf("expected HumidityMax 90.0, got %f", updated.HumidityMax)
	}
}

func TestPgAlertThresholdRepository_FindByID_NotFound(t *testing.T) {
	repo := setupPgAlertThresholdRepo(t)

	_, err := repo.FindByID(uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
