package repository

import (
	"sync"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// MemoryWorkRepository is a thread-safe in-memory implementation of WorkRepository.
type MemoryWorkRepository struct {
	mu    sync.RWMutex
	works map[uuid.UUID]*domain.Work
}

// NewMemoryWorkRepository creates a new in-memory work repository.
func NewMemoryWorkRepository() *MemoryWorkRepository {
	return &MemoryWorkRepository{
		works: make(map[uuid.UUID]*domain.Work),
	}
}

// FindByID retrieves a work by its ID. Returns ErrNotFound if it does not exist.
func (r *MemoryWorkRepository) FindByID(id uuid.UUID) (*domain.Work, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	work, ok := r.works[id]
	if !ok {
		return nil, ErrNotFound
	}
	copied := *work
	return &copied, nil
}

// FindAll retrieves all works from the repository.
func (r *MemoryWorkRepository) FindAll() ([]domain.Work, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Work, 0, len(r.works))
	for _, w := range r.works {
		result = append(result, *w)
	}
	return result, nil
}

// Create stores a new work. Returns ErrConflict if the ID already exists.
func (r *MemoryWorkRepository) Create(work *domain.Work) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.works[work.ID]; exists {
		return ErrConflict
	}
	copied := *work
	r.works[work.ID] = &copied
	return nil
}

// Update replaces an existing work. Returns ErrNotFound if not present.
func (r *MemoryWorkRepository) Update(work *domain.Work) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.works[work.ID]; !exists {
		return ErrNotFound
	}
	copied := *work
	r.works[work.ID] = &copied
	return nil
}

// Delete removes a work by ID. Returns ErrNotFound if not present.
func (r *MemoryWorkRepository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.works[id]; !exists {
		return ErrNotFound
	}
	delete(r.works, id)
	return nil
}

// Seed adds a work to the repository (for testing/bootstrapping).
func (r *MemoryWorkRepository) Seed(work *domain.Work) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *work
	r.works[work.ID] = &copied
}
