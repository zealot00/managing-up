package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

type mockExecutionRepo struct {
	execution     server.Execution
	skillVersion  server.SkillVersion
	updateCalls   []struct{ status, stepID string }
	createStepErr error
}

func (m *mockExecutionRepo) GetSkillVersionForExecution(skillID string) (server.SkillVersion, bool) {
	if m.skillVersion.SkillID == "" {
		return server.SkillVersion{}, false
	}
	return m.skillVersion, true
}

func (m *mockExecutionRepo) UpdateExecutionStatus(id, status, stepID string, endedAt *time.Time, durationMs *int64) error {
	m.updateCalls = append(m.updateCalls, struct{ status, stepID string }{status, stepID})
	return nil
}

func (m *mockExecutionRepo) CreateExecutionStep(step server.ExecutionStep) error {
	return m.createStepErr
}

func (m *mockExecutionRepo) GetExecutionForResume(id string) (server.Execution, bool) {
	if m.execution.ID == "" {
		return server.Execution{}, false
	}
	return m.execution, true
}

func TestEngine_Run_NotFound(t *testing.T) {
	repo := &mockExecutionRepo{}
	gw := NewToolGateway()
	engine := NewExecutionEngine(repo, gw)

	err := engine.Run(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for execution not found, got nil")
	}
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Errorf("expected ErrExecutionNotFound, got %v", err)
	}
}

func TestEngine_Run_SkillNotFound(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:      "exec-123",
			SkillID: "skill-456",
			Status:  "queued",
		},
	}
	gw := NewToolGateway()
	engine := NewExecutionEngine(repo, gw)

	err := engine.Run(context.Background(), "exec-123")
	if err == nil {
		t.Fatal("expected error for skill not found, got nil")
	}
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("expected ErrSkillNotFound, got %v", err)
	}
}

func TestEngine_Run_InvalidSpec(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:      "exec-123",
			SkillID: "skill-456",
			Status:  "queued",
		},
		skillVersion: server.SkillVersion{
			SkillID:  "skill-456",
			SpecYaml: "invalid: yaml: content: [",
		},
	}
	gw := NewToolGateway()
	engine := NewExecutionEngine(repo, gw)

	err := engine.Run(context.Background(), "exec-123")
	if err == nil {
		t.Fatal("expected error for invalid spec, got nil")
	}
}

func TestEngine_Run_ValidSpec_NoSteps(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:      "exec-123",
			SkillID: "skill-456",
			Status:  "queued",
		},
		skillVersion: server.SkillVersion{
			SkillID:  "skill-456",
			SpecYaml: "name: test\nversion: 1.0\nsteps:\n  - id: step1\n    type: tool\n    tool_ref: test-tool",
		},
	}
	gw := NewToolGateway()
	engine := NewExecutionEngine(repo, gw)

	err := engine.Run(context.Background(), "exec-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEngine_Run_SingleToolStep(t *testing.T) {
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:      "exec-123",
			SkillID: "skill-456",
			Status:  "queued",
		},
		skillVersion: server.SkillVersion{
			SkillID: "skill-456",
			SpecYaml: `name: test
version: "1.0"
steps:
  - id: step1
    type: tool
    tool_ref: test-tool
    with:
      key: value
`,
		},
	}
	gw := NewToolGateway()
	engine := NewExecutionEngine(repo, gw)

	err := engine.Run(context.Background(), "exec-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.updateCalls) < 2 {
		t.Fatalf("expected at least 2 update calls (running + succeeded), got %d", len(repo.updateCalls))
	}

	lastCall := repo.updateCalls[len(repo.updateCalls)-1]
	if lastCall.status != "succeeded" {
		t.Errorf("expected final status 'succeeded', got %q", lastCall.status)
	}
}
