// Package domain defines the core business entities for urushi-chronicle.
// These types model the lacquerware production workflow:
// works → process_steps → environment_readings → images.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Technique represents the type of lacquerware technique used.
type Technique string

const (
	TechniqueMakie      Technique = "makie"       // 蒔絵
	TechniqueRaden      Technique = "raden"       // 螺鈿
	TechniqueMakieRaden Technique = "makie_raden" // 蒔絵螺鈿
	TechniqueOther      Technique = "other"       // その他
)

// WorkStatus represents the current status of a work.
type WorkStatus string

const (
	WorkStatusInProgress WorkStatus = "in_progress"
	WorkStatusCompleted  WorkStatus = "completed"
	WorkStatusArchived   WorkStatus = "archived"
)

// Work represents a lacquerware piece being created or documented.
type Work struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Technique   Technique  `json:"technique"`
	Material    string     `json:"material,omitempty"`
	Status      WorkStatus `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Validate checks the work's fields for correctness.
func (w *Work) Validate() error {
	if w.Title == "" {
		return errors.New("work title must not be empty")
	}
	if len(w.Title) > 200 {
		return errors.New("work title must not exceed 200 characters")
	}
	if !isValidTechnique(w.Technique) {
		return errors.New("invalid technique")
	}
	if !isValidWorkStatus(w.Status) {
		return errors.New("invalid work status")
	}
	if w.CompletedAt != nil && w.CompletedAt.Before(w.StartedAt) {
		return errors.New("completed_at must be after started_at")
	}
	return nil
}

func isValidTechnique(t Technique) bool {
	switch t {
	case TechniqueMakie, TechniqueRaden, TechniqueMakieRaden, TechniqueOther:
		return true
	}
	return false
}

func isValidWorkStatus(s WorkStatus) bool {
	switch s {
	case WorkStatusInProgress, WorkStatusCompleted, WorkStatusArchived:
		return true
	}
	return false
}
