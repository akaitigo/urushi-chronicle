package domain_test

import (
	"testing"
	"time"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	"github.com/google/uuid"
)

func validWork() domain.Work {
	return domain.Work{
		ID:        uuid.New(),
		Title:     "棗（なつめ）蒔絵",
		Technique: domain.TechniqueMakie,
		Status:    domain.WorkStatusInProgress,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestWorkValidate_Valid(t *testing.T) {
	w := validWork()
	if err := w.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestWorkValidate_EmptyTitle(t *testing.T) {
	w := validWork()
	w.Title = ""
	if err := w.Validate(); err == nil {
		t.Error("expected error for empty title")
	}
}

func TestWorkValidate_TitleTooLong(t *testing.T) {
	w := validWork()
	w.Title = string(make([]byte, 201))
	if err := w.Validate(); err == nil {
		t.Error("expected error for title exceeding 200 characters")
	}
}

func TestWorkValidate_InvalidTechnique(t *testing.T) {
	w := validWork()
	w.Technique = "invalid"
	if err := w.Validate(); err == nil {
		t.Error("expected error for invalid technique")
	}
}

func TestWorkValidate_InvalidStatus(t *testing.T) {
	w := validWork()
	w.Status = "unknown"
	if err := w.Validate(); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestWorkValidate_CompletedBeforeStarted(t *testing.T) {
	w := validWork()
	past := w.StartedAt.Add(-1 * time.Hour)
	w.CompletedAt = &past
	if err := w.Validate(); err == nil {
		t.Error("expected error when completed_at is before started_at")
	}
}

func TestWorkValidate_AllTechniques(t *testing.T) {
	techniques := []domain.Technique{
		domain.TechniqueMakie,
		domain.TechniqueRaden,
		domain.TechniqueMakieRaden,
		domain.TechniqueOther,
	}
	for _, tech := range techniques {
		w := validWork()
		w.Technique = tech
		if err := w.Validate(); err != nil {
			t.Errorf("technique %s should be valid, got %v", tech, err)
		}
	}
}

func TestWorkValidate_AllStatuses(t *testing.T) {
	statuses := []domain.WorkStatus{
		domain.WorkStatusInProgress,
		domain.WorkStatusCompleted,
		domain.WorkStatusArchived,
	}
	for _, s := range statuses {
		w := validWork()
		w.Status = s
		if err := w.Validate(); err != nil {
			t.Errorf("status %s should be valid, got %v", s, err)
		}
	}
}
