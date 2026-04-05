package repository

import (
	"context"
	"fmt"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgEnvironmentRepository is a PostgreSQL+TimescaleDB implementation of EnvironmentRepository.
type PgEnvironmentRepository struct {
	pool *pgxpool.Pool
}

// NewPgEnvironmentRepository creates a new PostgreSQL environment repository.
func NewPgEnvironmentRepository(pool *pgxpool.Pool) *PgEnvironmentRepository {
	return &PgEnvironmentRepository{pool: pool}
}

// Store inserts a single environment reading into the TimescaleDB hypertable.
func (r *PgEnvironmentRepository) Store(reading *domain.EnvironmentReading) error {
	ctx := context.Background()
	query := `
		INSERT INTO environment_readings (time, sensor_id, location, temperature, humidity, work_id, process_step_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		reading.Time,
		reading.SensorID,
		reading.Location,
		reading.Temperature,
		reading.Humidity,
		reading.WorkID,
		reading.ProcessStepID,
	)
	if err != nil {
		return fmt.Errorf("failed to store environment reading: %w", err)
	}
	return nil
}

// FindBySensorID retrieves readings for a given sensor, ordered by time descending.
// If limit <= 0, a default limit of 1000 is applied to prevent unbounded queries.
func (r *PgEnvironmentRepository) FindBySensorID(sensorID string, limit int) ([]domain.EnvironmentReading, error) {
	ctx := context.Background()
	if limit <= 0 {
		limit = 1000
	}

	query := `
		SELECT time, sensor_id, location, temperature, humidity, work_id, process_step_id
		FROM environment_readings
		WHERE sensor_id = $1
		ORDER BY time DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, sensorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment readings: %w", err)
	}
	defer rows.Close()

	var readings []domain.EnvironmentReading
	for rows.Next() {
		var reading domain.EnvironmentReading
		if err := rows.Scan(
			&reading.Time,
			&reading.SensorID,
			&reading.Location,
			&reading.Temperature,
			&reading.Humidity,
			&reading.WorkID,
			&reading.ProcessStepID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan environment reading: %w", err)
		}
		readings = append(readings, reading)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment readings: %w", err)
	}

	return readings, nil
}
