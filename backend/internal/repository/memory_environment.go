package repository

import (
	"sort"
	"sync"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

// MemoryEnvironmentRepository is a thread-safe in-memory implementation of EnvironmentRepository.
type MemoryEnvironmentRepository struct {
	mu       sync.RWMutex
	readings []domain.EnvironmentReading
}

// NewMemoryEnvironmentRepository creates a new in-memory environment repository.
func NewMemoryEnvironmentRepository() *MemoryEnvironmentRepository {
	return &MemoryEnvironmentRepository{
		readings: make([]domain.EnvironmentReading, 0),
	}
}

// Store saves a single environment reading.
func (r *MemoryEnvironmentRepository) Store(reading *domain.EnvironmentReading) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *reading
	r.readings = append(r.readings, copied)
	return nil
}

// FindBySensorID retrieves readings for a given sensor, ordered by time descending.
// If limit <= 0, all readings are returned.
func (r *MemoryEnvironmentRepository) FindBySensorID(sensorID string, limit int) ([]domain.EnvironmentReading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.EnvironmentReading
	for i := range r.readings {
		if r.readings[i].SensorID == sensorID {
			result = append(result, r.readings[i])
		}
	}

	// Sort by time descending (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.After(result[j].Time)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}
