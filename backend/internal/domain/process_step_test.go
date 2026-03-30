package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

func validProcessStep() domain.ProcessStep {
	return domain.ProcessStep{
		ID:        uuid.New(),
		WorkID:    uuid.New(),
		Name:      "下塗り一回目",
		StepOrder: 1,
		Category:  domain.StepCategoryShitanuri,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestProcessStepValidate_Valid(t *testing.T) {
	ps := validProcessStep()
	if err := ps.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestProcessStepValidate_EmptyName(t *testing.T) {
	ps := validProcessStep()
	ps.Name = ""
	if err := ps.Validate(); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestProcessStepValidate_NameTooLong(t *testing.T) {
	ps := validProcessStep()
	ps.Name = strings.Repeat("a", 101)
	if err := ps.Validate(); err == nil {
		t.Error("expected error for name exceeding 100 characters")
	}
}

func TestProcessStepValidate_NilWorkID(t *testing.T) {
	ps := validProcessStep()
	ps.WorkID = uuid.Nil
	if err := ps.Validate(); err == nil {
		t.Error("expected error for nil work_id")
	}
}

func TestProcessStepValidate_ZeroStepOrder(t *testing.T) {
	ps := validProcessStep()
	ps.StepOrder = 0
	if err := ps.Validate(); err == nil {
		t.Error("expected error for step_order < 1")
	}
}

func TestProcessStepValidate_InvalidCategory(t *testing.T) {
	ps := validProcessStep()
	ps.Category = "invalid"
	if err := ps.Validate(); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestProcessStepValidate_CompletedBeforeStarted(t *testing.T) {
	ps := validProcessStep()
	past := ps.StartedAt.Add(-1 * time.Hour)
	ps.CompletedAt = &past
	if err := ps.Validate(); err == nil {
		t.Error("expected error when completed_at is before started_at")
	}
}

func TestProcessStepValidate_AllCategories(t *testing.T) {
	categories := []domain.StepCategory{
		domain.StepCategoryShitanuri,
		domain.StepCategoryNakanuri,
		domain.StepCategoryUwanuri,
		domain.StepCategoryMakie,
		domain.StepCategoryRaden,
		domain.StepCategoryTogidashi,
		domain.StepCategoryRoiro,
		domain.StepCategoryOther,
	}
	for _, c := range categories {
		ps := validProcessStep()
		ps.Category = c
		if err := ps.Validate(); err != nil {
			t.Errorf("category %s should be valid, got %v", c, err)
		}
	}
}
