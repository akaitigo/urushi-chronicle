package repository_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func TestMemoryWorkRepository_FindByID(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := &domain.Work{
		ID:        uuid.New(),
		Title:     "テスト作品",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	repo.Seed(work)

	found, err := repo.FindByID(work.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != work.Title {
		t.Errorf("expected title %q, got %q", work.Title, found.Title)
	}
}

func TestMemoryWorkRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	_, err := repo.FindByID(uuid.New())
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryWorkRepository_Seed_ReturnsIndependentCopy(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := &domain.Work{
		ID:        uuid.New(),
		Title:     "オリジナル",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	repo.Seed(work)

	// Mutate original
	work.Title = "変更後"

	found, _ := repo.FindByID(work.ID)
	if found.Title != "オリジナル" {
		t.Errorf("seed should store a copy, got %q", found.Title)
	}
}
