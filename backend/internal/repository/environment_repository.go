package repository

import (
	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

// EnvironmentRepository defines the interface for environment reading persistence.
type EnvironmentRepository interface {
	// Store saves a single environment reading.
	Store(reading *domain.EnvironmentReading) error
	// FindBySensorID retrieves all readings for a given sensor, ordered by time descending.
	FindBySensorID(sensorID string, limit int) ([]domain.EnvironmentReading, error)
}
