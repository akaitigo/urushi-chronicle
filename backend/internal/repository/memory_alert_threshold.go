package repository

import (
	"sync"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// MemoryAlertThresholdRepository is a thread-safe in-memory implementation of AlertThresholdRepository.
type MemoryAlertThresholdRepository struct {
	mu         sync.RWMutex
	thresholds map[uuid.UUID]*domain.AlertThreshold
}

// NewMemoryAlertThresholdRepository creates a new in-memory alert threshold repository.
func NewMemoryAlertThresholdRepository() *MemoryAlertThresholdRepository {
	return &MemoryAlertThresholdRepository{
		thresholds: make(map[uuid.UUID]*domain.AlertThreshold),
	}
}

// Create stores a new alert threshold.
func (r *MemoryAlertThresholdRepository) Create(threshold *domain.AlertThreshold) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *threshold
	r.thresholds[threshold.ID] = &copied
	return nil
}

// FindByID retrieves an alert threshold by its ID.
func (r *MemoryAlertThresholdRepository) FindByID(id uuid.UUID) (*domain.AlertThreshold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	threshold, ok := r.thresholds[id]
	if !ok {
		return nil, ErrNotFound
	}
	copied := *threshold
	return &copied, nil
}

// FindBySensorID retrieves all enabled alert thresholds for a given sensor.
func (r *MemoryAlertThresholdRepository) FindBySensorID(sensorID string) ([]domain.AlertThreshold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AlertThreshold
	for _, t := range r.thresholds {
		if t.SensorID == sensorID && t.Enabled {
			result = append(result, *t)
		}
	}
	return result, nil
}

// FindAllEnabled retrieves all enabled alert thresholds.
func (r *MemoryAlertThresholdRepository) FindAllEnabled() ([]domain.AlertThreshold, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.AlertThreshold
	for _, t := range r.thresholds {
		if t.Enabled {
			result = append(result, *t)
		}
	}
	return result, nil
}

// Update replaces an existing alert threshold.
func (r *MemoryAlertThresholdRepository) Update(threshold *domain.AlertThreshold) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.thresholds[threshold.ID]; !exists {
		return ErrNotFound
	}
	copied := *threshold
	r.thresholds[threshold.ID] = &copied
	return nil
}

// Delete removes an alert threshold by its ID.
func (r *MemoryAlertThresholdRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.thresholds[id]; !exists {
		return ErrNotFound
	}
	delete(r.thresholds, id)
	return nil
}

// Seed adds an alert threshold to the repository (for testing/bootstrapping).
func (r *MemoryAlertThresholdRepository) Seed(threshold *domain.AlertThreshold) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *threshold
	r.thresholds[threshold.ID] = &copied
}
