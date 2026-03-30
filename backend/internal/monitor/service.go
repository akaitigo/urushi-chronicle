// Package monitor provides the environment monitoring service that integrates
// sensor data ingestion, persistence, and alert evaluation.
package monitor

import (
	"log"

	"github.com/akaitigo/urushi-chronicle/internal/alert"
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/akaitigo/urushi-chronicle/internal/repository"
)

// Service orchestrates environment reading ingestion and alert evaluation.
type Service struct {
	envRepo       repository.EnvironmentRepository
	thresholdRepo repository.AlertThresholdRepository
	notifier      alert.Notifier
	logger        *log.Logger
}

// NewService creates a new monitoring Service with the given dependencies.
func NewService(
	envRepo repository.EnvironmentRepository,
	thresholdRepo repository.AlertThresholdRepository,
	notifier alert.Notifier,
	logger *log.Logger,
) *Service {
	return &Service{
		envRepo:       envRepo,
		thresholdRepo: thresholdRepo,
		notifier:      notifier,
		logger:        logger,
	}
}

// ProcessReading stores a reading and evaluates it against all applicable alert thresholds.
// If any threshold is exceeded, an alert notification is sent.
// Storage errors are returned immediately. Alert notification errors are logged but do not
// prevent the reading from being stored.
func (s *Service) ProcessReading(reading domain.EnvironmentReading) error {
	// Validate the reading
	if err := reading.Validate(); err != nil {
		return err
	}

	// Store the reading
	if err := s.envRepo.Store(&reading); err != nil {
		return err
	}

	// Look up applicable thresholds for this sensor
	thresholds, err := s.thresholdRepo.FindBySensorID(reading.SensorID)
	if err != nil {
		s.logger.Printf("warning: failed to lookup thresholds for sensor %s: %v", reading.SensorID, err)
		return nil
	}

	// Evaluate each threshold
	for i := range thresholds {
		if thresholds[i].IsOutOfRange(reading) {
			s.logger.Printf("alert: sensor %s at %s exceeded threshold %s",
				reading.SensorID, reading.Location, thresholds[i].ID)
			if notifyErr := s.notifier.Notify(reading, thresholds[i]); notifyErr != nil {
				s.logger.Printf("error: failed to send alert notification: %v", notifyErr)
			}
		}
	}

	return nil
}
