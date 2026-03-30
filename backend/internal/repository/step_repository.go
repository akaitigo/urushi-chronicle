// Package repository defines data access interfaces and in-memory implementations
// for the urushi-chronicle domain entities.
package repository

import (
	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// StepRepository defines the interface for process step persistence.
type StepRepository interface {
	Create(step *domain.ProcessStep) error
	FindByID(workID, stepID uuid.UUID) (*domain.ProcessStep, error)
	FindByWorkID(workID uuid.UUID) ([]domain.ProcessStep, error)
	Update(step *domain.ProcessStep) error
	Delete(workID, stepID uuid.UUID) error
}

// WorkRepository defines the interface for work persistence.
type WorkRepository interface {
	FindByID(id uuid.UUID) (*domain.Work, error)
}
