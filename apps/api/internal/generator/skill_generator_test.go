package generator

import (
	"context"
	"errors"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// mockLLMClient is a mock implementation of llm.Client for testing
type mockLLMClient struct {
	resp *llm.Response
	err  error
}

func (m *mockLLMClient) Generate(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
	return m.resp, m.err
}

func (m *mockLLMClient) Provider() llm.Provider { return llm.ProviderOpenAI }
func (m *mockLLMClient) Model() llm.Model       { return llm.ModelGPT4o }

// TestValidate_Success tests that a valid spec passes validation
func TestValidate_Success(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name:      "test_skill",
		Version:   "1.0",
		RiskLevel: "medium",
		Steps: []Step{
			{ID: "step_1", Type: "tool", ToolRef: "http_request"},
		},
	}
	if err := g.validate(spec); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestValidate_EmptyName tests that empty name fails validation
func TestValidate_EmptyName(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "",
		Steps: []Step{
			{ID: "step_1", Type: "tool", ToolRef: "http_request"},
		},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for empty name")
	} else if err.Error() != "name is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_EmptySteps tests that empty steps fails validation
func TestValidate_EmptySteps(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name:  "test_skill",
		Steps: []Step{},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for empty steps")
	} else if err.Error() != "at least one step is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_DuplicateStepIDs tests that duplicate step IDs fail validation
func TestValidate_DuplicateStepIDs(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "test_skill",
		Steps: []Step{
			{ID: "step_1", Type: "tool", ToolRef: "http_request"},
			{ID: "step_1", Type: "tool", ToolRef: "http_request"},
		},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for duplicate step IDs")
	} else if err.Error() != "duplicate step id: step_1" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_ToolStepWithoutToolRef tests that tool step without tool_ref fails validation
func TestValidate_ToolStepWithoutToolRef(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "test_skill",
		Steps: []Step{
			{ID: "step_1", Type: "tool", ToolRef: ""},
		},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for tool step without tool_ref")
	} else if err.Error() != "tool_ref is required for tool step step_1" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_ApprovalStepWithoutApproverGroup tests that approval step without approver_group fails validation
func TestValidate_ApprovalStepWithoutApproverGroup(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "test_skill",
		Steps: []Step{
			{ID: "step_1", Type: "approval", ApproverGroup: ""},
		},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for approval step without approver_group")
	} else if err.Error() != "approver_group is required for approval step step_1" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_EmptyStepID tests that empty step ID fails validation
func TestValidate_EmptyStepID(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "test_skill",
		Steps: []Step{
			{ID: "", Type: "tool", ToolRef: "http_request"},
		},
	}
	if err := g.validate(spec); err == nil {
		t.Error("expected error for empty step ID")
	} else if err.Error() != "step id is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestValidate_DefaultValues tests that default values are set for version and risk_level
func TestValidate_DefaultValues(t *testing.T) {
	g := &Generator{}
	spec := &SkillSpec{
		Name: "test_skill",
		Steps: []Step{
			{ID: "step_1", Type: "tool", ToolRef: "http_request"},
		},
	}
	if err := g.validate(spec); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if spec.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", spec.Version)
	}
	if spec.RiskLevel != "medium" {
		t.Errorf("expected risk_level medium, got %s", spec.RiskLevel)
	}
}

// TestBuildSOPPrompt_Empty tests that empty SOP text returns empty prompt
func TestBuildSOPPrompt_Empty(t *testing.T) {
	prompt := buildSOPPrompt("", "")
	expected := "Convert this SOP to a skill specification YAML:\n\n"
	if prompt != expected {
		t.Errorf("expected %q, got %q", expected, prompt)
	}
}

// TestBuildSOPPrompt_WithDescription tests that description is included
func TestBuildSOPPrompt_WithDescription(t *testing.T) {
	prompt := buildSOPPrompt("Do something", "Additional context here")
	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !contains(prompt, "Do something") {
		t.Error("expected prompt to contain SOP text")
	}
	if !contains(prompt, "Additional context: Additional context here") {
		t.Error("expected prompt to contain description")
	}
}

// TestSystemPrompt_NotEmpty tests that system prompt is not empty
func TestSystemPrompt_NotEmpty(t *testing.T) {
	prompt := systemPrompt()
	if prompt == "" {
		t.Error("expected non-empty system prompt")
	}
	if !contains(prompt, "skill specification generator") {
		t.Error("expected system prompt to contain skill specification generator")
	}
	if !contains(prompt, "YAML") {
		t.Error("expected system prompt to contain YAML")
	}
}

// TestGenerateFromSOP_Success tests successful generation from SOP
func TestGenerateFromSOP_Success(t *testing.T) {
	validYAML := `name: test_skill
version: "1.0"
risk_level: medium
description: A test skill
steps:
  - id: step_1
    type: tool
    tool_ref: http_request
`
	mockClient := &mockLLMClient{
		resp: &llm.Response{
			Content:      validYAML,
			Provider:     llm.ProviderOpenAI,
			Model:        llm.ModelGPT4o,
			Usage:        llm.Usage{InputTokens: 10, OutputTokens: 20, TotalTokens: 30},
			FinishReason: "stop",
		},
		err: nil,
	}
	g := NewGenerator(mockClient)
	req := GenerateRequest{
		SOPText:     "Do something useful",
		Description: "Test description",
	}
	resp, err := g.GenerateFromSOP(context.Background(), req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Spec == nil {
		t.Fatal("expected spec, got nil")
	}
	if resp.Spec.Name != "test_skill" {
		t.Errorf("expected name test_skill, got %s", resp.Spec.Name)
	}
	if resp.Provider != llm.ProviderOpenAI {
		t.Errorf("expected provider openai, got %s", resp.Provider)
	}
	if resp.Model != llm.ModelGPT4o {
		t.Errorf("expected model gpt-4o, got %s", resp.Model)
	}
	if resp.Usage == nil {
		t.Error("expected usage, got nil")
	}
}

// TestGenerateFromSOP_LLMError tests that LLM error is propagated
func TestGenerateFromSOP_LLMError(t *testing.T) {
	mockClient := &mockLLMClient{
		resp: nil,
		err:  errors.New("llm connection failed"),
	}
	g := NewGenerator(mockClient)
	req := GenerateRequest{
		SOPText: "Do something",
	}
	_, err := g.GenerateFromSOP(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "llm call failed") {
		t.Errorf("expected error to contain 'llm call failed', got: %v", err)
	}
}

// TestGenerateFromSOP_InvalidYAML tests that invalid YAML returns parsing error
func TestGenerateFromSOP_InvalidYAML(t *testing.T) {
	invalidYAML := `name: test_skill
version: "1.0"
steps:
  - id: step_1
    type: tool
`
	mockClient := &mockLLMClient{
		resp: &llm.Response{
			Content:      invalidYAML,
			Provider:     llm.ProviderOpenAI,
			Model:        llm.ModelGPT4o,
			Usage:        llm.Usage{},
			FinishReason: "stop",
		},
		err: nil,
	}
	g := NewGenerator(mockClient)
	req := GenerateRequest{
		SOPText: "Do something",
	}
	_, err := g.GenerateFromSOP(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "invalid skill spec") {
		t.Errorf("expected error to contain 'invalid skill spec', got: %v", err)
	}
}

// TestGenerateFromSOP_ValidationError tests that YAML missing required fields returns validation error
func TestGenerateFromSOP_ValidationError(t *testing.T) {
	// Missing tool_ref for tool step - will fail parser validation
	invalidYAML := `name: test_skill
version: "1.0"
steps:
  - id: step_1
    type: tool
`
	mockClient := &mockLLMClient{
		resp: &llm.Response{
			Content:      invalidYAML,
			Provider:     llm.ProviderOpenAI,
			Model:        llm.ModelGPT4o,
			Usage:        llm.Usage{},
			FinishReason: "stop",
		},
		err: nil,
	}
	g := NewGenerator(mockClient)
	req := GenerateRequest{
		SOPText: "Do something",
	}
	_, err := g.GenerateFromSOP(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should fail because tool_ref is missing for tool step
	if !contains(err.Error(), "invalid skill spec") {
		t.Errorf("expected error to contain 'invalid skill spec', got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
