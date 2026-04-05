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

func setupPgWorkRepo(t *testing.T) *repository.PgWorkRepository {
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

	return repository.NewPgWorkRepository(pool)
}

func newPgTestWork(title string) *domain.Work {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &domain.Work{
		ID:          uuid.New(),
		Title:       title,
		Description: "テスト用作品",
		Technique:   domain.TechniqueMakie,
		Material:    "欅",
		Status:      domain.WorkStatusInProgress,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestPgWorkRepository_CRUD(t *testing.T) {
	repo := setupPgWorkRepo(t)

	work := newPgTestWork("PGテスト蒔絵")
	t.Cleanup(func() {
		// Best-effort cleanup
		_ = repo.Delete(work.ID)
	})

	// Create
	if err := repo.Create(work); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(work.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != "PGテスト蒔絵" {
		t.Errorf("expected title 'PGテスト蒔絵', got %q", found.Title)
	}

	// Update
	found.Title = "更新後タイトル"
	found.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(found); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	updated, err := repo.FindByID(work.ID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if updated.Title != "更新後タイトル" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}

	// FindAll
	works, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(works) == 0 {
		t.Error("expected at least 1 work in FindAll")
	}

	// Delete
	if err := repo.Delete(work.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.FindByID(work.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPgWorkRepository_Create_Conflict(t *testing.T) {
	repo := setupPgWorkRepo(t)

	work := newPgTestWork("重複テスト")
	t.Cleanup(func() {
		_ = repo.Delete(work.ID)
	})

	if err := repo.Create(work); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	err := repo.Create(work)
	if !errors.Is(err, repository.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate insert, got %v", err)
	}
}

func TestPgWorkRepository_FindByID_NotFound(t *testing.T) {
	repo := setupPgWorkRepo(t)

	_, err := repo.FindByID(uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
