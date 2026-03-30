package repository_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func newStep(workID uuid.UUID, order int) *domain.ProcessStep {
	return &domain.ProcessStep{
		ID:        uuid.New(),
		WorkID:    workID,
		Name:      "テスト工程",
		StepOrder: order,
		Category:  domain.StepCategoryShitanuri,
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func TestMemoryStepRepository_CreateAndFindByID(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()
	step := newStep(workID, 1)

	if err := repo.Create(step); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(workID, step.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != step.Name {
		t.Errorf("expected name %q, got %q", step.Name, found.Name)
	}
}

func TestMemoryStepRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	_, err := repo.FindByID(uuid.New(), uuid.New())
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStepRepository_FindByWorkID_Ordered(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()

	// Create in reverse order
	for _, order := range []int{3, 1, 2} {
		step := newStep(workID, order)
		if err := repo.Create(step); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	steps, err := repo.FindByWorkID(workID)
	if err != nil {
		t.Fatalf("FindByWorkID failed: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	for i, step := range steps {
		if step.StepOrder != i+1 {
			t.Errorf("step[%d] order = %d, want %d", i, step.StepOrder, i+1)
		}
	}
}

func TestMemoryStepRepository_FindByWorkID_EmptyResult(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	steps, err := repo.FindByWorkID(uuid.New())
	if err != nil {
		t.Fatalf("FindByWorkID failed: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(steps))
	}
}

func TestMemoryStepRepository_Create_StepOrderConflict(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()

	step1 := newStep(workID, 1)
	if err := repo.Create(step1); err != nil {
		t.Fatalf("Create step1 failed: %v", err)
	}

	step2 := newStep(workID, 1) // same order
	if err := repo.Create(step2); err != repository.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestMemoryStepRepository_Update(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()
	step := newStep(workID, 1)

	if err := repo.Create(step); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	step.Name = "更新された工程"
	if err := repo.Update(step); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.FindByID(workID, step.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "更新された工程" {
		t.Errorf("expected updated name, got %q", found.Name)
	}
}

func TestMemoryStepRepository_Update_NotFound(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	step := newStep(uuid.New(), 1)
	if err := repo.Update(step); err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStepRepository_Update_StepOrderConflict(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()

	step1 := newStep(workID, 1)
	step2 := newStep(workID, 2)
	_ = repo.Create(step1)
	_ = repo.Create(step2)

	// Try to change step2's order to 1
	step2.StepOrder = 1
	if err := repo.Update(step2); err != repository.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestMemoryStepRepository_Delete(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	workID := uuid.New()
	step := newStep(workID, 1)

	_ = repo.Create(step)

	if err := repo.Delete(workID, step.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID(workID, step.ID)
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStepRepository_Delete_NotFound(t *testing.T) {
	repo := repository.NewMemoryStepRepository()
	if err := repo.Delete(uuid.New(), uuid.New()); err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
