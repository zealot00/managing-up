package engine

import (
	"errors"
	"testing"
)

func TestParse_ValidSpec(t *testing.T) {
	parser := NewSkillSpecParser()
	yamlContent := `
name: test-skill
version: "1.0.0"
description: A test skill
steps:
  - id: step1
    type: tool
    tool_ref: test-tool
`
	spec, err := parser.Parse(yamlContent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if spec.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got %q", spec.Name)
	}
	if spec.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", spec.Version)
	}
	if len(spec.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(spec.Steps))
	}
	if spec.Steps[0].ID != "step1" {
		t.Errorf("expected step id 'step1', got %q", spec.Steps[0].ID)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	parser := NewSkillSpecParser()
	yamlContent := `invalid: yaml: content: [`

	_, err := parser.Parse(yamlContent)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestParse_EmptySpec(t *testing.T) {
	parser := NewSkillSpecParser()

	_, err := parser.Parse("")
	if err == nil {
		t.Fatal("expected error for empty YAML, got nil")
	}
	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestParse_MissingName(t *testing.T) {
	parser := NewSkillSpecParser()
	yamlContent := `
version: "1.0.0"
steps:
  - id: step1
    type: tool
    tool_ref: test-tool
`
	_, err := parser.Parse(yamlContent)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestParse_MissingVersion(t *testing.T) {
	parser := NewSkillSpecParser()
	yamlContent := `
name: test-skill
steps:
  - id: step1
    type: tool
    tool_ref: test-tool
`
	_, err := parser.Parse(yamlContent)
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}
	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestParse_NoSteps(t *testing.T) {
	parser := NewSkillSpecParser()
	yamlContent := `
name: test-skill
version: "1.0.0"
`
	_, err := parser.Parse(yamlContent)
	if err == nil {
		t.Fatal("expected error for no steps, got nil")
	}
	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestValidateStepOrder_Valid(t *testing.T) {
	parser := NewSkillSpecParser()
	spec := &SkillSpec{
		Name:    "test",
		Version: "1.0",
		Steps: []Step{
			{ID: "step1", Type: "tool", ToolRef: "tool1"},
			{ID: "step2", Type: "tool", ToolRef: "tool2"},
			{ID: "step3", Type: "approval", ApproverGroup: "admins"},
		},
	}

	err := parser.ValidateStepOrder(spec)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateStepOrder_Invalid(t *testing.T) {
	parser := NewSkillSpecParser()
	spec := &SkillSpec{
		Name:    "test",
		Version: "1.0",
		Steps: []Step{
			{ID: "step1", Type: "tool", ToolRef: "tool1"},
			{ID: "step1", Type: "tool", ToolRef: "tool2"},
		},
	}

	err := parser.ValidateStepOrder(spec)
	if err == nil {
		t.Fatal("expected error for duplicate step id, got nil")
	}
}
