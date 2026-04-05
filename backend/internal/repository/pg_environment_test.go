package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPgEnvRepo(t *testing.T) *repository.PgEnvironmentRepository {
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

	// Clean test data
	_, err = pool.Exec(context.Background(), "DELETE FROM environment_readings WHERE sensor_id LIKE 'test-%'")
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	return repository.NewPgEnvironmentRepository(pool)
}

func TestPgEnvironmentRepository_StoreAndFind(t *testing.T) {
	repo := setupPgEnvRepo(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	reading := &domain.EnvironmentReading{
		Time:        now,
		SensorID:    "test-sensor-001",
		Location:    "urushi_buro",
		Temperature: 25.0,
		Humidity:    75.0,
	}

	if err := repo.Store(reading); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	results, err := repo.FindBySensorID("test-sensor-001", 10)
	if err != nil {
		t.Fatalf("FindBySensorID failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 reading, got 0")
	}

	got := results[0]
	if got.SensorID != "test-sensor-001" {
		t.Errorf("expected sensor_id 'test-sensor-001', got %q", got.SensorID)
	}
	if got.Temperature != 25.0 {
		t.Errorf("expected temperature 25.0, got %f", got.Temperature)
	}
}

func TestPgEnvironmentRepository_FindBySensorID_Empty(t *testing.T) {
	repo := setupPgEnvRepo(t)

	results, err := repo.FindBySensorID("test-nonexistent-sensor", 10)
	if err != nil {
		t.Fatalf("FindBySensorID failed: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil for nonexistent sensor, got %d results", len(results))
	}
}
