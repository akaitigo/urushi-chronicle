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

func TestMemoryWorkRepository_FindAll_Empty(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	works, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(works) != 0 {
		t.Errorf("expected 0 works, got %d", len(works))
	}
}

func TestMemoryWorkRepository_FindAll_WithData(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		repo.Seed(&domain.Work{
			ID:        uuid.New(),
			Title:     "作品",
			Technique: domain.TechniqueMakie,
			Status:    domain.WorkStatusInProgress,
			StartedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	works, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(works) != 3 {
		t.Errorf("expected 3 works, got %d", len(works))
	}
}

func TestMemoryWorkRepository_FindAll_ReturnsIndependentCopies(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	now := time.Now().UTC()
	repo.Seed(&domain.Work{
		ID:        uuid.New(),
		Title:     "オリジナル",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	})

	works, _ := repo.FindAll()
	works[0].Title = "変更後"

	worksAgain, _ := repo.FindAll()
	if worksAgain[0].Title != "オリジナル" {
		t.Errorf("FindAll should return copies, got %q", worksAgain[0].Title)
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
