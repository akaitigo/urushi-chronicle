package repository_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
	"github.com/google/uuid"
)

func newTestWork(title string) *domain.Work {
	now := time.Now().UTC()
	return &domain.Work{
		ID:        uuid.New(),
		Title:     title,
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMemoryWorkRepository_FindByID(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("テスト作品")
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
	for i := 0; i < 3; i++ {
		repo.Seed(newTestWork("作品"))
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
	repo.Seed(newTestWork("オリジナル"))

	works, _ := repo.FindAll()
	works[0].Title = "変更後"

	worksAgain, _ := repo.FindAll()
	if worksAgain[0].Title != "オリジナル" {
		t.Errorf("FindAll should return copies, got %q", worksAgain[0].Title)
	}
}

func TestMemoryWorkRepository_Seed_ReturnsIndependentCopy(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("オリジナル")
	repo.Seed(work)

	// Mutate original
	work.Title = "変更後"

	found, _ := repo.FindByID(work.ID)
	if found.Title != "オリジナル" {
		t.Errorf("seed should store a copy, got %q", found.Title)
	}
}

func TestMemoryWorkRepository_Create(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("新規作品")

	if err := repo.Create(work); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(work.ID)
	if err != nil {
		t.Fatalf("FindByID after Create failed: %v", err)
	}
	if found.Title != "新規作品" {
		t.Errorf("expected title '新規作品', got %q", found.Title)
	}
}

func TestMemoryWorkRepository_Create_Conflict(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("作品")
	repo.Seed(work)

	err := repo.Create(work)
	if err != repository.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestMemoryWorkRepository_Create_StoresIndependentCopy(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("オリジナル")

	if err := repo.Create(work); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	work.Title = "変更後"

	found, _ := repo.FindByID(work.ID)
	if found.Title != "オリジナル" {
		t.Errorf("Create should store a copy, got %q", found.Title)
	}
}

func TestMemoryWorkRepository_Update(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("旧タイトル")
	repo.Seed(work)

	work.Title = "新タイトル"
	if err := repo.Update(work); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(work.ID)
	if found.Title != "新タイトル" {
		t.Errorf("expected '新タイトル', got %q", found.Title)
	}
}

func TestMemoryWorkRepository_Update_NotFound(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("存在しない")

	err := repo.Update(work)
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryWorkRepository_Update_StoresIndependentCopy(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("オリジナル")
	repo.Seed(work)

	work.Title = "更新後"
	if err := repo.Update(work); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	work.Title = "さらに変更"

	found, _ := repo.FindByID(work.ID)
	if found.Title != "更新後" {
		t.Errorf("Update should store a copy, got %q", found.Title)
	}
}

func TestMemoryWorkRepository_Delete(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()
	work := newTestWork("削除対象")
	repo.Seed(work)

	if err := repo.Delete(work.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID(work.ID)
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryWorkRepository_Delete_NotFound(t *testing.T) {
	repo := repository.NewMemoryWorkRepository()

	err := repo.Delete(uuid.New())
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
