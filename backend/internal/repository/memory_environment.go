package repository

import (
	"sort"
	"sync"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
)

// maxReadings is the maximum number of readings the in-memory store will hold.
// Once this limit is reached, the oldest readings are overwritten using a ring buffer.
const maxReadings = 100_000

// MemoryEnvironmentRepository is a thread-safe in-memory implementation of EnvironmentRepository.
// It uses a ring buffer to cap memory usage at maxReadings entries.
type MemoryEnvironmentRepository struct {
	mu       sync.RWMutex
	readings []domain.EnvironmentReading
	head     int  // next write position in the ring buffer
	full     bool // true once the buffer has wrapped around at least once
}

// NewMemoryEnvironmentRepository creates a new in-memory environment repository.
func NewMemoryEnvironmentRepository() *MemoryEnvironmentRepository {
	return &MemoryEnvironmentRepository{
		readings: make([]domain.EnvironmentReading, maxReadings),
	}
}

// Store saves a single environment reading using a ring buffer.
// When the buffer is full, the oldest entry is overwritten.
func (r *MemoryEnvironmentRepository) Store(reading *domain.EnvironmentReading) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.readings[r.head] = *reading
	r.head++
	if r.head >= maxReadings {
		r.head = 0
		r.full = true
	}
	return nil
}

// count returns the number of valid entries in the ring buffer.
// Caller must hold at least a read lock.
func (r *MemoryEnvironmentRepository) count() int {
	if r.full {
		return maxReadings
	}
	return r.head
}

// FindBySensorID retrieves readings for a given sensor, ordered by time descending.
// If limit <= 0, all readings are returned.
func (r *MemoryEnvironmentRepository) FindBySensorID(sensorID string, limit int) ([]domain.EnvironmentReading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := r.count()
	var result []domain.EnvironmentReading
	for i := 0; i < n; i++ {
		idx := i
		if r.full {
			idx = (r.head + i) % maxReadings
		}
		if r.readings[idx].SensorID == sensorID {
			result = append(result, r.readings[idx])
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
