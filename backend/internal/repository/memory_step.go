package repository

import (
	"errors"
	"sort"
	"sync"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a uniqueness constraint is violated.
var ErrConflict = errors.New("step_order conflict: another step already uses this order")

// MemoryStepRepository is a thread-safe in-memory implementation of StepRepository.
type MemoryStepRepository struct {
	mu    sync.RWMutex
	steps map[uuid.UUID]map[uuid.UUID]*domain.ProcessStep // workID -> stepID -> step
}

// NewMemoryStepRepository creates a new in-memory step repository.
func NewMemoryStepRepository() *MemoryStepRepository {
	return &MemoryStepRepository{
		steps: make(map[uuid.UUID]map[uuid.UUID]*domain.ProcessStep),
	}
}

// Create stores a new process step. Returns ErrConflict if step_order is already taken for the work.
func (r *MemoryStepRepository) Create(step *domain.ProcessStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	workSteps, ok := r.steps[step.WorkID]
	if !ok {
		workSteps = make(map[uuid.UUID]*domain.ProcessStep)
		r.steps[step.WorkID] = workSteps
	}

	// Check step_order uniqueness within the work
	for _, existing := range workSteps {
		if existing.StepOrder == step.StepOrder {
			return ErrConflict
		}
	}

	copied := *step
	workSteps[step.ID] = &copied
	return nil
}

// FindByID retrieves a single process step by work ID and step ID.
func (r *MemoryStepRepository) FindByID(workID, stepID uuid.UUID) (*domain.ProcessStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workSteps, ok := r.steps[workID]
	if !ok {
		return nil, ErrNotFound
	}
	step, ok := workSteps[stepID]
	if !ok {
		return nil, ErrNotFound
	}
	copied := *step
	return &copied, nil
}

// FindByWorkID retrieves all process steps for a work, ordered by step_order.
func (r *MemoryStepRepository) FindByWorkID(workID uuid.UUID) ([]domain.ProcessStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workSteps, ok := r.steps[workID]
	if !ok {
		return []domain.ProcessStep{}, nil
	}

	result := make([]domain.ProcessStep, 0, len(workSteps))
	for _, step := range workSteps {
		result = append(result, *step)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StepOrder < result[j].StepOrder
	})
	return result, nil
}

// Update replaces an existing process step. Returns ErrNotFound if it does not exist.
func (r *MemoryStepRepository) Update(step *domain.ProcessStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	workSteps, ok := r.steps[step.WorkID]
	if !ok {
		return ErrNotFound
	}
	if _, exists := workSteps[step.ID]; !exists {
		return ErrNotFound
	}

	// Check step_order uniqueness (excluding current step)
	for id, existing := range workSteps {
		if id != step.ID && existing.StepOrder == step.StepOrder {
			return ErrConflict
		}
	}

	copied := *step
	workSteps[step.ID] = &copied
	return nil
}

// Delete removes a process step. Returns ErrNotFound if it does not exist.
func (r *MemoryStepRepository) Delete(workID, stepID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	workSteps, ok := r.steps[workID]
	if !ok {
		return ErrNotFound
	}
	if _, exists := workSteps[stepID]; !exists {
		return ErrNotFound
	}
	delete(workSteps, stepID)
	return nil
}
