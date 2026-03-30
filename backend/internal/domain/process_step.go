package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// StepCategory represents the type of production process step.
type StepCategory string

const (
	StepCategoryShitanuri StepCategory = "shitanuri" // 下塗り
	StepCategoryNakanuri  StepCategory = "nakanuri"  // 中塗り
	StepCategoryUwanuri   StepCategory = "uwanuri"   // 上塗り
	StepCategoryMakie     StepCategory = "makie"     // 蒔絵
	StepCategoryRaden     StepCategory = "raden"     // 螺鈿
	StepCategoryTogidashi StepCategory = "togidashi" // 研ぎ出し
	StepCategoryRoiro     StepCategory = "roiro"     // 呂色仕上げ
	StepCategoryOther     StepCategory = "other"     // その他
)

// ProcessStep represents a single step in the lacquerware production process.
type ProcessStep struct {
	ID            uuid.UUID    `json:"id"`
	WorkID        uuid.UUID    `json:"work_id"`
	Name          string       `json:"name"`
	Description   string       `json:"description,omitempty"`
	StepOrder     int          `json:"step_order"`
	Category      StepCategory `json:"category"`
	MaterialsUsed []string     `json:"materials_used,omitempty"`
	Notes         string       `json:"notes,omitempty"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// MaxStepNameLength is the maximum allowed length for a process step name.
const MaxStepNameLength = 100

// Validate checks the process step's fields for correctness.
func (ps *ProcessStep) Validate() error {
	if ps.Name == "" {
		return errors.New("process step name must not be empty")
	}
	if len(ps.Name) > MaxStepNameLength {
		return errors.New("process step name must not exceed 100 characters")
	}
	if ps.WorkID == uuid.Nil {
		return errors.New("work_id must not be empty")
	}
	if ps.StepOrder < 1 {
		return errors.New("step_order must be at least 1")
	}
	if !isValidStepCategory(ps.Category) {
		return errors.New("invalid step category")
	}
	if ps.CompletedAt != nil && ps.CompletedAt.Before(ps.StartedAt) {
		return errors.New("completed_at must be after started_at")
	}
	return nil
}

func isValidStepCategory(c StepCategory) bool {
	switch c {
	case StepCategoryShitanuri, StepCategoryNakanuri, StepCategoryUwanuri,
		StepCategoryMakie, StepCategoryRaden, StepCategoryTogidashi,
		StepCategoryRoiro, StepCategoryOther:
		return true
	}
	return false
}
