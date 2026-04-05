package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgWorkRepository is a PostgreSQL implementation of WorkRepository.
type PgWorkRepository struct {
	pool *pgxpool.Pool
}

// NewPgWorkRepository creates a new PostgreSQL work repository.
func NewPgWorkRepository(pool *pgxpool.Pool) *PgWorkRepository {
	return &PgWorkRepository{pool: pool}
}

// FindByID retrieves a work by its ID.
func (r *PgWorkRepository) FindByID(id uuid.UUID) (*domain.Work, error) {
	ctx := context.Background()
	query := `
		SELECT id, title, description, technique, material, status,
		       started_at, completed_at, created_at, updated_at
		FROM works
		WHERE id = $1
	`
	var w domain.Work
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.Title, &w.Description, &w.Technique, &w.Material, &w.Status,
		&w.StartedAt, &w.CompletedAt, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find work: %w", err)
	}
	return &w, nil
}

// FindAll retrieves all works ordered by creation time descending.
func (r *PgWorkRepository) FindAll() ([]domain.Work, error) {
	ctx := context.Background()
	query := `
		SELECT id, title, description, technique, material, status,
		       started_at, completed_at, created_at, updated_at
		FROM works
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list works: %w", err)
	}
	defer rows.Close()

	var works []domain.Work
	for rows.Next() {
		var w domain.Work
		if err := rows.Scan(
			&w.ID, &w.Title, &w.Description, &w.Technique, &w.Material, &w.Status,
			&w.StartedAt, &w.CompletedAt, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan work: %w", err)
		}
		works = append(works, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating works: %w", err)
	}

	if works == nil {
		works = []domain.Work{}
	}
	return works, nil
}

// Create inserts a new work. Returns ErrConflict if the ID already exists.
func (r *PgWorkRepository) Create(work *domain.Work) error {
	ctx := context.Background()
	query := `
		INSERT INTO works (id, title, description, technique, material, status,
		                   started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		work.ID, work.Title, work.Description, work.Technique, work.Material, work.Status,
		work.StartedAt, work.CompletedAt, work.CreatedAt, work.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("failed to create work: %w", err)
	}
	return nil
}

// Update replaces an existing work. Returns ErrNotFound if not present.
func (r *PgWorkRepository) Update(work *domain.Work) error {
	ctx := context.Background()
	query := `
		UPDATE works
		SET title = $2, description = $3, technique = $4, material = $5, status = $6,
		    started_at = $7, completed_at = $8, updated_at = $9
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query,
		work.ID, work.Title, work.Description, work.Technique, work.Material, work.Status,
		work.StartedAt, work.CompletedAt, work.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a work by ID. Returns ErrNotFound if not present.
func (r *PgWorkRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	query := `DELETE FROM works WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
