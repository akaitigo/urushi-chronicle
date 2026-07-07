// Package repository defines data access interfaces and in-memory implementations
// for the urushi-chronicle domain entities.
package repository

import (
	"context"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

// StepRepository defines the interface for process step persistence.
type StepRepository interface {
	Create(ctx context.Context, step *domain.ProcessStep) error
	FindByID(ctx context.Context, workID, stepID uuid.UUID) (*domain.ProcessStep, error)
	FindByWorkID(ctx context.Context, workID uuid.UUID) ([]domain.ProcessStep, error)
	Update(ctx context.Context, step *domain.ProcessStep) error
	Delete(ctx context.Context, workID, stepID uuid.UUID) error
	DeleteByWorkID(ctx context.Context, workID uuid.UUID) error
}

// WorkRepository defines the interface for work persistence.
type WorkRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	FindAll(ctx context.Context) ([]domain.Work, error)
	Create(ctx context.Context, work *domain.Work) error
	Update(ctx context.Context, work *domain.Work) error
	Delete(ctx context.Context, id uuid.UUID) error
}
