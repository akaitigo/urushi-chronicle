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

func setupPgStepRepo(t *testing.T) (*repository.PgStepRepository, *repository.PgWorkRepository) {
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

	return repository.NewPgStepRepository(pool), repository.NewPgWorkRepository(pool)
}

func newPgTestStep(workID uuid.UUID, name string, order int) *domain.ProcessStep {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &domain.ProcessStep{
		ID:            uuid.New(),
		WorkID:        workID,
		Name:          name,
		Description:   "テスト工程",
		StepOrder:     order,
		Category:      domain.StepCategoryShitanuri,
		MaterialsUsed: []string{"漆", "砥の粉"},
		Notes:         "テスト用ノート",
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestPgStepRepository_CRUD(t *testing.T) {
	stepRepo, workRepo := setupPgStepRepo(t)

	// Create parent work first
	work := newPgTestWork("工程テスト親作品")
	t.Cleanup(func() {
		_ = stepRepo.DeleteByWorkID(work.ID)
		_ = workRepo.Delete(work.ID)
	})
	if err := workRepo.Create(work); err != nil {
		t.Fatalf("Create work failed: %v", err)
	}

	step := newPgTestStep(work.ID, "下塗り", 1)

	// Create
	if err := stepRepo.Create(step); err != nil {
		t.Fatalf("Create step failed: %v", err)
	}

	// FindByID
	found, err := stepRepo.FindByID(work.ID, step.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "下塗り" {
		t.Errorf("expected name '下塗り', got %q", found.Name)
	}
	if len(found.MaterialsUsed) != 2 {
		t.Errorf("expected 2 materials, got %d", len(found.MaterialsUsed))
	}

	// FindByWorkID
	steps, err := stepRepo.FindByWorkID(work.ID)
	if err != nil {
		t.Fatalf("FindByWorkID failed: %v", err)
	}
	if len(steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(steps))
	}

	// Update
	found.Name = "更新下塗り"
	found.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := stepRepo.Update(found); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Delete
	if err := stepRepo.Delete(work.ID, step.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = stepRepo.FindByID(work.ID, step.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPgStepRepository_Create_Conflict(t *testing.T) {
	stepRepo, workRepo := setupPgStepRepo(t)

	work := newPgTestWork("重複工程テスト")
	t.Cleanup(func() {
		_ = stepRepo.DeleteByWorkID(work.ID)
		_ = workRepo.Delete(work.ID)
	})
	if err := workRepo.Create(work); err != nil {
		t.Fatalf("Create work failed: %v", err)
	}

	step1 := newPgTestStep(work.ID, "工程1", 1)
	if err := stepRepo.Create(step1); err != nil {
		t.Fatalf("Create step1 failed: %v", err)
	}

	step2 := newPgTestStep(work.ID, "工程2", 1) // same step_order
	err := stepRepo.Create(step2)
	if !errors.Is(err, repository.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate step_order, got %v", err)
	}
}

func TestPgStepRepository_DeleteByWorkID(t *testing.T) {
	stepRepo, workRepo := setupPgStepRepo(t)

	work := newPgTestWork("一括削除テスト")
	t.Cleanup(func() {
		_ = workRepo.Delete(work.ID)
	})
	if err := workRepo.Create(work); err != nil {
		t.Fatalf("Create work failed: %v", err)
	}

	for i := 1; i <= 3; i++ {
		step := newPgTestStep(work.ID, "工程", i)
		if err := stepRepo.Create(step); err != nil {
			t.Fatalf("Create step %d failed: %v", i, err)
		}
	}

	if err := stepRepo.DeleteByWorkID(work.ID); err != nil {
		t.Fatalf("DeleteByWorkID failed: %v", err)
	}

	steps, err := stepRepo.FindByWorkID(work.ID)
	if err != nil {
		t.Fatalf("FindByWorkID after delete failed: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps after DeleteByWorkID, got %d", len(steps))
	}
}
