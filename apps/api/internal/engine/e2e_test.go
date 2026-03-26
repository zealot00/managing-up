package engine

import (
	"context"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/engine/tool/builtin"
	"github.com/zealot/managing-up/apps/api/internal/server"
)

func TestE2E_CalculatorExecution(t *testing.T) {
	// Create mock repository with a skill version that has a calculator step
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:      "exec-e2e-123",
			SkillID: "skill-calc-456",
			Status:  "queued",
		},
		skillVersion: server.SkillVersion{
			SkillID: "skill-calc-456",
			SpecYaml: `name: test-calc
version: "1.0"
steps:
  - id: calc-step
    type: tool
    tool_ref: calculator
    with:
      operation: add
      a: "2"
      b: "3"
`,
		},
	}

	// Create tool gateway with calculator registered
	gw := NewToolGateway()
	if err := gw.Register(builtin.NewCalculator(), nil); err != nil {
		t.Fatalf("failed to register calculator: %v", err)
	}

	// Create execution engine
	engine := NewExecutionEngine(repo, gw)

	// Run execution
	err := engine.Run(context.Background(), "exec-e2e-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify execution succeeded
	if len(repo.updateCalls) < 2 {
		t.Fatalf("expected at least 2 update calls (running + succeeded), got %d", len(repo.updateCalls))
	}

	lastCall := repo.updateCalls[len(repo.updateCalls)-1]
	if lastCall.status != "succeeded" {
		t.Errorf("expected final status 'succeeded', got %q", lastCall.status)
	}
}
