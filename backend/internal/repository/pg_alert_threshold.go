package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgAlertThresholdRepository is a PostgreSQL implementation of AlertThresholdRepository.
type PgAlertThresholdRepository struct {
	pool *pgxpool.Pool
}

// NewPgAlertThresholdRepository creates a new PostgreSQL alert threshold repository.
func NewPgAlertThresholdRepository(pool *pgxpool.Pool) *PgAlertThresholdRepository {
	return &PgAlertThresholdRepository{pool: pool}
}

// Create stores a new alert threshold.
func (r *PgAlertThresholdRepository) Create(threshold *domain.AlertThreshold) error {
	ctx := context.Background()
	query := `
		INSERT INTO alert_thresholds
			(id, sensor_id, temperature_min, temperature_max, humidity_min, humidity_max,
			 enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		threshold.ID, threshold.SensorID,
		threshold.TemperatureMin, threshold.TemperatureMax,
		threshold.HumidityMin, threshold.HumidityMax,
		threshold.Enabled, threshold.CreatedAt, threshold.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create alert threshold: %w", err)
	}
	return nil
}

// FindByID retrieves an alert threshold by its ID.
func (r *PgAlertThresholdRepository) FindByID(id uuid.UUID) (*domain.AlertThreshold, error) {
	ctx := context.Background()
	query := `
		SELECT id, sensor_id, temperature_min, temperature_max, humidity_min, humidity_max,
		       enabled, created_at, updated_at
		FROM alert_thresholds
		WHERE id = $1
	`
	var t domain.AlertThreshold
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.SensorID,
		&t.TemperatureMin, &t.TemperatureMax,
		&t.HumidityMin, &t.HumidityMax,
		&t.Enabled, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find alert threshold: %w", err)
	}
	return &t, nil
}

// FindBySensorID retrieves all enabled alert thresholds for a given sensor.
func (r *PgAlertThresholdRepository) FindBySensorID(sensorID string) ([]domain.AlertThreshold, error) {
	ctx := context.Background()
	query := `
		SELECT id, sensor_id, temperature_min, temperature_max, humidity_min, humidity_max,
		       enabled, created_at, updated_at
		FROM alert_thresholds
		WHERE sensor_id = $1 AND enabled = TRUE
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, sensorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert thresholds: %w", err)
	}
	defer rows.Close()

	var thresholds []domain.AlertThreshold
	for rows.Next() {
		var t domain.AlertThreshold
		if err := rows.Scan(
			&t.ID, &t.SensorID,
			&t.TemperatureMin, &t.TemperatureMax,
			&t.HumidityMin, &t.HumidityMax,
			&t.Enabled, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert threshold: %w", err)
		}
		thresholds = append(thresholds, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert thresholds: %w", err)
	}

	return thresholds, nil
}

// FindAllEnabled retrieves all enabled alert thresholds.
func (r *PgAlertThresholdRepository) FindAllEnabled() ([]domain.AlertThreshold, error) {
	ctx := context.Background()
	query := `
		SELECT id, sensor_id, temperature_min, temperature_max, humidity_min, humidity_max,
		       enabled, created_at, updated_at
		FROM alert_thresholds
		WHERE enabled = TRUE
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled alert thresholds: %w", err)
	}
	defer rows.Close()

	var thresholds []domain.AlertThreshold
	for rows.Next() {
		var t domain.AlertThreshold
		if err := rows.Scan(
			&t.ID, &t.SensorID,
			&t.TemperatureMin, &t.TemperatureMax,
			&t.HumidityMin, &t.HumidityMax,
			&t.Enabled, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert threshold: %w", err)
		}
		thresholds = append(thresholds, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert thresholds: %w", err)
	}

	return thresholds, nil
}

// Update replaces an existing alert threshold.
func (r *PgAlertThresholdRepository) Update(threshold *domain.AlertThreshold) error {
	ctx := context.Background()
	query := `
		UPDATE alert_thresholds
		SET sensor_id = $2, temperature_min = $3, temperature_max = $4,
		    humidity_min = $5, humidity_max = $6, enabled = $7, updated_at = $8
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query,
		threshold.ID, threshold.SensorID,
		threshold.TemperatureMin, threshold.TemperatureMax,
		threshold.HumidityMin, threshold.HumidityMax,
		threshold.Enabled, threshold.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update alert threshold: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FindAndUpdate atomically retrieves and updates a threshold using a database transaction.
func (r *PgAlertThresholdRepository) FindAndUpdate(id uuid.UUID, updateFn func(existing *domain.AlertThreshold) (*domain.AlertThreshold, error)) (*domain.AlertThreshold, error) {
	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a safe no-op
	}()

	// SELECT FOR UPDATE to lock the row
	query := `
		SELECT id, sensor_id, temperature_min, temperature_max, humidity_min, humidity_max,
		       enabled, created_at, updated_at
		FROM alert_thresholds
		WHERE id = $1
		FOR UPDATE
	`
	var t domain.AlertThreshold
	err = tx.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.SensorID,
		&t.TemperatureMin, &t.TemperatureMax,
		&t.HumidityMin, &t.HumidityMax,
		&t.Enabled, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find alert threshold for update: %w", err)
	}

	updated, err := updateFn(&t)
	if err != nil {
		return nil, err
	}

	updateQuery := `
		UPDATE alert_thresholds
		SET sensor_id = $2, temperature_min = $3, temperature_max = $4,
		    humidity_min = $5, humidity_max = $6, enabled = $7, updated_at = $8
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, updateQuery,
		updated.ID, updated.SensorID,
		updated.TemperatureMin, updated.TemperatureMax,
		updated.HumidityMin, updated.HumidityMax,
		updated.Enabled, updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update alert threshold in transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return updated, nil
}

// Delete removes an alert threshold by its ID.
func (r *PgAlertThresholdRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	query := `DELETE FROM alert_thresholds WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert threshold: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
