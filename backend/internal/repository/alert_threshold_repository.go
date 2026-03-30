package repository

import (
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// AlertThresholdRepository defines the interface for alert threshold persistence.
type AlertThresholdRepository interface {
	// Create stores a new alert threshold.
	Create(threshold *domain.AlertThreshold) error
	// FindByID retrieves an alert threshold by its ID.
	FindByID(id uuid.UUID) (*domain.AlertThreshold, error)
	// FindBySensorID retrieves all enabled alert thresholds for a given sensor.
	FindBySensorID(sensorID string) ([]domain.AlertThreshold, error)
	// FindAllEnabled retrieves all enabled alert thresholds.
	FindAllEnabled() ([]domain.AlertThreshold, error)
	// Update replaces an existing alert threshold.
	Update(threshold *domain.AlertThreshold) error
	// Delete removes an alert threshold by its ID.
	Delete(id uuid.UUID) error
}
