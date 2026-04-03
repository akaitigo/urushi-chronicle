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
// A sensor_id index (map[string][]int) accelerates FindBySensorID from O(n) to O(k)
// where k is the number of readings for the given sensor.
type MemoryEnvironmentRepository struct {
	mu          sync.RWMutex
	readings    []domain.EnvironmentReading
	head        int              // next write position in the ring buffer
	full        bool             // true once the buffer has wrapped around at least once
	sensorIndex map[string][]int // sensor_id -> ring buffer indices
}

// NewMemoryEnvironmentRepository creates a new in-memory environment repository.
func NewMemoryEnvironmentRepository() *MemoryEnvironmentRepository {
	return &MemoryEnvironmentRepository{
		readings:    make([]domain.EnvironmentReading, maxReadings),
		sensorIndex: make(map[string][]int),
	}
}

// Store saves a single environment reading using a ring buffer.
// When the buffer is full, the oldest entry is overwritten and its index entry is removed.
func (r *MemoryEnvironmentRepository) Store(reading *domain.EnvironmentReading) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pos := r.head

	// If overwriting an existing entry, remove the old index entry
	if r.full {
		oldSensorID := r.readings[pos].SensorID
		if oldSensorID != "" {
			r.removeFromIndex(oldSensorID, pos)
		}
	}

	r.readings[pos] = *reading

	// Add to sensor index
	r.sensorIndex[reading.SensorID] = append(r.sensorIndex[reading.SensorID], pos)

	r.head++
	if r.head >= maxReadings {
		r.head = 0
		r.full = true
	}
	return nil
}

// removeFromIndex removes a specific position from the sensor's index list.
// Caller must hold the write lock.
func (r *MemoryEnvironmentRepository) removeFromIndex(sensorID string, pos int) {
	indices := r.sensorIndex[sensorID]
	for i, idx := range indices {
		if idx == pos {
			// Remove by swapping with last element
			indices[i] = indices[len(indices)-1]
			indices = indices[:len(indices)-1]
			break
		}
	}
	if len(indices) == 0 {
		delete(r.sensorIndex, sensorID)
	} else {
		r.sensorIndex[sensorID] = indices
	}
}

// FindBySensorID retrieves readings for a given sensor, ordered by time descending.
// If limit <= 0, all readings are returned.
// Uses the sensor_id index for O(k) lookup instead of O(n) full scan.
func (r *MemoryEnvironmentRepository) FindBySensorID(sensorID string, limit int) ([]domain.EnvironmentReading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	indices := r.sensorIndex[sensorID]
	if len(indices) == 0 {
		return nil, nil
	}

	result := make([]domain.EnvironmentReading, 0, len(indices))
	for _, idx := range indices {
		result = append(result, r.readings[idx])
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
