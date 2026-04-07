package engine

import (
	"context"
	"os"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

func TestE2E_GeneratedSkill(t *testing.T) {
	if os.Getenv("SKIP_E2E") == "1" {
		t.Skip("skipping E2E test")
	}
	data, err := os.ReadFile("/tmp/skill-test/skill.schema.json")
	if err != nil {
		t.Skipf("skipping E2E test: %v", err)
	}

	skillYaml := string(data)

	// Create mock repo with the generated skill
	repo := &mockExecutionRepo{
		execution: server.Execution{
			ID:          "exec-test-generated",
			SkillID:     "generated-skill",
			Status:      "queued",
			TriggeredBy: "test",
			Input:       map[string]any{},
		},
		skillVersion: server.SkillVersion{
			SkillID:  "generated-skill",
			SpecYaml: skillYaml,
		},
	}

	// Create tool gateway (no registration needed - uses HTTP mock)
	gw := NewToolGateway()

	// Create execution engine
	eng := NewExecutionEngine(repo, gw)

	// Run execution
	err = eng.Run(context.Background(), "exec-test-generated")
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify execution succeeded
	if len(repo.updateCalls) < 2 {
		t.Fatalf("expected at least 2 update calls, got %d", len(repo.updateCalls))
	}

	lastCall := repo.updateCalls[len(repo.updateCalls)-1]
	if lastCall.status != "succeeded" {
		t.Errorf("expected final status 'succeeded', got %q", lastCall.status)
	}

	t.Logf("Generated skill executed successfully!")
	t.Logf("  Total update calls: %d", len(repo.updateCalls))
	t.Logf("  Final status: %s", lastCall.status)
}
