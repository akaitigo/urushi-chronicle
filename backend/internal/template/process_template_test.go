package template_test

import (
	"testing"

	"github.com/akaitigo/urushi-chronicle/internal/domain"
	tmpl "github.com/akaitigo/urushi-chronicle/internal/template"
)

func TestDefaultWorkflow_HasFiveSteps(t *testing.T) {
	wf := tmpl.DefaultWorkflow()
	if len(wf.Steps) != 5 {
		t.Errorf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestDefaultWorkflow_CorrectOrder(t *testing.T) {
	wf := tmpl.DefaultWorkflow()
	expectedCategories := []domain.StepCategory{
		domain.StepCategoryShitanuri,
		domain.StepCategoryNakanuri,
		domain.StepCategoryUwanuri,
		domain.StepCategoryMakie,
		domain.StepCategoryRaden,
	}
	for i, step := range wf.Steps {
		if step.Order != i+1 {
			t.Errorf("step[%d] order = %d, want %d", i, step.Order, i+1)
		}
		if step.Category != expectedCategories[i] {
			t.Errorf("step[%d] category = %s, want %s", i, step.Category, expectedCategories[i])
		}
	}
}

func TestMakieWorkflow_HasSevenSteps(t *testing.T) {
	wf := tmpl.MakieWorkflow()
	if len(wf.Steps) != 7 {
		t.Errorf("expected 7 steps, got %d", len(wf.Steps))
	}
	if wf.Name != "makie" {
		t.Errorf("expected name 'makie', got %q", wf.Name)
	}
}

func TestRadenWorkflow_HasSevenSteps(t *testing.T) {
	wf := tmpl.RadenWorkflow()
	if len(wf.Steps) != 7 {
		t.Errorf("expected 7 steps, got %d", len(wf.Steps))
	}
	if wf.Name != "raden" {
		t.Errorf("expected name 'raden', got %q", wf.Name)
	}
}

func TestGetWorkflow_ReturnsCorrectTemplate(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"makie", "makie"},
		{"raden", "raden"},
		{"standard", "standard"},
		{"unknown", "standard"}, // defaults to standard
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := tmpl.GetWorkflow(tt.name)
			if wf.Name != tt.expected {
				t.Errorf("GetWorkflow(%q) returned %q, want %q", tt.name, wf.Name, tt.expected)
			}
		})
	}
}

func TestAllWorkflows_ReturnsThree(t *testing.T) {
	all := tmpl.AllWorkflows()
	if len(all) != 3 {
		t.Errorf("expected 3 workflows, got %d", len(all))
	}
	names := make(map[string]bool)
	for _, wf := range all {
		names[wf.Name] = true
	}
	for _, expected := range []string{"standard", "makie", "raden"} {
		if !names[expected] {
			t.Errorf("missing workflow %q", expected)
		}
	}
}
