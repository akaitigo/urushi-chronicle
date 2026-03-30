// Package template provides predefined process step templates for
// common lacquerware production workflows.
package template

import "github.com/akaitigo/urushi-chronicle/internal/domain"

// StepTemplate represents a single step in a predefined production workflow.
type StepTemplate struct {
	Name     string              `json:"name"`
	Category domain.StepCategory `json:"category"`
	Order    int                 `json:"order"`
}

// WorkflowTemplate represents a complete predefined production workflow.
type WorkflowTemplate struct {
	Name  string         `json:"name"`
	Steps []StepTemplate `json:"steps"`
}

// DefaultWorkflow returns the standard lacquerware production workflow:
// 下塗り → 中塗り → 上塗り → 蒔絵 → 螺鈿
func DefaultWorkflow() WorkflowTemplate {
	return WorkflowTemplate{
		Name: "standard",
		Steps: []StepTemplate{
			{Name: "下塗り", Category: domain.StepCategoryShitanuri, Order: 1},
			{Name: "中塗り", Category: domain.StepCategoryNakanuri, Order: 2},
			{Name: "上塗り", Category: domain.StepCategoryUwanuri, Order: 3},
			{Name: "蒔絵", Category: domain.StepCategoryMakie, Order: 4},
			{Name: "螺鈿", Category: domain.StepCategoryRaden, Order: 5},
		},
	}
}

// MakieWorkflow returns a makie-focused workflow.
func MakieWorkflow() WorkflowTemplate {
	return WorkflowTemplate{
		Name: "makie",
		Steps: []StepTemplate{
			{Name: "下塗り", Category: domain.StepCategoryShitanuri, Order: 1},
			{Name: "中塗り", Category: domain.StepCategoryNakanuri, Order: 2},
			{Name: "上塗り", Category: domain.StepCategoryUwanuri, Order: 3},
			{Name: "蒔絵 — 下絵", Category: domain.StepCategoryMakie, Order: 4},
			{Name: "蒔絵 — 金粉蒔き", Category: domain.StepCategoryMakie, Order: 5},
			{Name: "研ぎ出し", Category: domain.StepCategoryTogidashi, Order: 6},
			{Name: "呂色仕上げ", Category: domain.StepCategoryRoiro, Order: 7},
		},
	}
}

// RadenWorkflow returns a raden-focused workflow.
func RadenWorkflow() WorkflowTemplate {
	return WorkflowTemplate{
		Name: "raden",
		Steps: []StepTemplate{
			{Name: "下塗り", Category: domain.StepCategoryShitanuri, Order: 1},
			{Name: "中塗り", Category: domain.StepCategoryNakanuri, Order: 2},
			{Name: "上塗り", Category: domain.StepCategoryUwanuri, Order: 3},
			{Name: "螺鈿 — 貝片配置", Category: domain.StepCategoryRaden, Order: 4},
			{Name: "螺鈿 — 固定・埋め", Category: domain.StepCategoryRaden, Order: 5},
			{Name: "研ぎ出し", Category: domain.StepCategoryTogidashi, Order: 6},
			{Name: "呂色仕上げ", Category: domain.StepCategoryRoiro, Order: 7},
		},
	}
}

// GetWorkflow returns a workflow template by name. Returns the default workflow if not found.
func GetWorkflow(name string) WorkflowTemplate {
	switch name {
	case "makie":
		return MakieWorkflow()
	case "raden":
		return RadenWorkflow()
	default:
		return DefaultWorkflow()
	}
}

// AllWorkflows returns all available workflow templates.
func AllWorkflows() []WorkflowTemplate {
	return []WorkflowTemplate{
		DefaultWorkflow(),
		MakieWorkflow(),
		RadenWorkflow(),
	}
}
