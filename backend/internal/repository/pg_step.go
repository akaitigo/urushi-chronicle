package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStepRepository is a PostgreSQL implementation of StepRepository.
type PgStepRepository struct {
	pool *pgxpool.Pool
}

// NewPgStepRepository creates a new PostgreSQL step repository.
func NewPgStepRepository(pool *pgxpool.Pool) *PgStepRepository {
	return &PgStepRepository{pool: pool}
}

// Create inserts a new process step. Returns ErrConflict if step_order is already taken for the work.
func (r *PgStepRepository) Create(step *domain.ProcessStep) error {
	ctx := context.Background()

	materialsJSON, err := json.Marshal(step.MaterialsUsed)
	if err != nil {
		return fmt.Errorf("failed to marshal materials_used: %w", err)
	}

	query := `
		INSERT INTO process_steps
			(id, work_id, name, description, step_order, category,
			 materials_used, notes, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = r.pool.Exec(ctx, query,
		step.ID, step.WorkID, step.Name, step.Description, step.StepOrder, step.Category,
		materialsJSON, step.Notes, step.StartedAt, step.CompletedAt, step.CreatedAt, step.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("failed to create process step: %w", err)
	}
	return nil
}

// FindByID retrieves a single process step by work ID and step ID.
func (r *PgStepRepository) FindByID(workID, stepID uuid.UUID) (*domain.ProcessStep, error) {
	ctx := context.Background()
	query := `
		SELECT id, work_id, name, description, step_order, category,
		       materials_used, notes, started_at, completed_at, created_at, updated_at
		FROM process_steps
		WHERE work_id = $1 AND id = $2
	`
	var s domain.ProcessStep
	var materialsJSON []byte
	err := r.pool.QueryRow(ctx, query, workID, stepID).Scan(
		&s.ID, &s.WorkID, &s.Name, &s.Description, &s.StepOrder, &s.Category,
		&materialsJSON, &s.Notes, &s.StartedAt, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find process step: %w", err)
	}

	if err := json.Unmarshal(materialsJSON, &s.MaterialsUsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal materials_used: %w", err)
	}

	return &s, nil
}

// FindByWorkID retrieves all process steps for a work, ordered by step_order.
func (r *PgStepRepository) FindByWorkID(workID uuid.UUID) ([]domain.ProcessStep, error) {
	ctx := context.Background()
	query := `
		SELECT id, work_id, name, description, step_order, category,
		       materials_used, notes, started_at, completed_at, created_at, updated_at
		FROM process_steps
		WHERE work_id = $1
		ORDER BY step_order ASC
	`
	rows, err := r.pool.Query(ctx, query, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to list process steps: %w", err)
	}
	defer rows.Close()

	var steps []domain.ProcessStep
	for rows.Next() {
		var s domain.ProcessStep
		var materialsJSON []byte
		if err := rows.Scan(
			&s.ID, &s.WorkID, &s.Name, &s.Description, &s.StepOrder, &s.Category,
			&materialsJSON, &s.Notes, &s.StartedAt, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan process step: %w", err)
		}
		if err := json.Unmarshal(materialsJSON, &s.MaterialsUsed); err != nil {
			return nil, fmt.Errorf("failed to unmarshal materials_used: %w", err)
		}
		steps = append(steps, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating process steps: %w", err)
	}

	if steps == nil {
		steps = []domain.ProcessStep{}
	}
	return steps, nil
}

// Update replaces an existing process step. Returns ErrNotFound if it does not exist.
// Returns ErrConflict if step_order uniqueness is violated.
func (r *PgStepRepository) Update(step *domain.ProcessStep) error {
	ctx := context.Background()

	materialsJSON, err := json.Marshal(step.MaterialsUsed)
	if err != nil {
		return fmt.Errorf("failed to marshal materials_used: %w", err)
	}

	query := `
		UPDATE process_steps
		SET name = $3, description = $4, step_order = $5, category = $6,
		    materials_used = $7, notes = $8, started_at = $9, completed_at = $10, updated_at = $11
		WHERE id = $1 AND work_id = $2
	`
	tag, err := r.pool.Exec(ctx, query,
		step.ID, step.WorkID, step.Name, step.Description, step.StepOrder, step.Category,
		materialsJSON, step.Notes, step.StartedAt, step.CompletedAt, step.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("failed to update process step: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a process step. Returns ErrNotFound if it does not exist.
func (r *PgStepRepository) Delete(workID, stepID uuid.UUID) error {
	ctx := context.Background()
	query := `DELETE FROM process_steps WHERE work_id = $1 AND id = $2`
	tag, err := r.pool.Exec(ctx, query, workID, stepID)
	if err != nil {
		return fmt.Errorf("failed to delete process step: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteByWorkID removes all process steps associated with a given work.
// Returns nil if the work has no steps (idempotent).
func (r *PgStepRepository) DeleteByWorkID(workID uuid.UUID) error {
	ctx := context.Background()
	query := `DELETE FROM process_steps WHERE work_id = $1`
	_, err := r.pool.Exec(ctx, query, workID)
	if err != nil {
		return fmt.Errorf("failed to delete process steps by work_id: %w", err)
	}
	return nil
}
