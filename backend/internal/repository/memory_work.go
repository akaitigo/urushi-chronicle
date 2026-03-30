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

// Seed adds a work to the repository (for testing/bootstrapping).
func (r *MemoryWorkRepository) Seed(work *domain.Work) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *work
	r.works[work.ID] = &copied
}
