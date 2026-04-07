package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestToolGateway_Invoke_Success(t *testing.T) {
	gw := NewToolGateway()
	inv := GatewayToolInvocation{
		ExecutionID:    "exec-123",
		StepID:         "step-1",
		ToolRef:        "test-tool",
		Input:          map[string]any{"key": "value"},
		TimeoutSeconds: 30,
	}

	result, err := gw.Invoke(context.Background(), inv)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "succeeded" {
		t.Errorf("expected status 'succeeded', got %q", result.Status)
	}
	if result.Output == nil {
		t.Fatal("expected output to be set")
	}
	if result.Output["tool_ref"] != "test-tool" {
		t.Errorf("expected tool_ref 'test-tool', got %v", result.Output["tool_ref"])
	}
	if result.Output["execution_id"] != "exec-123" {
		t.Errorf("expected execution_id 'exec-123', got %v", result.Output["execution_id"])
	}
}

func TestToolGateway_Invoke_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	gw := NewToolGateway()
	inv := GatewayToolInvocation{
		ExecutionID:    "exec-123",
		StepID:         "step-1",
		ToolRef:        "test-tool",
		Input:          map[string]any{},
		TimeoutSeconds: 30,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := gw.Invoke(ctx, inv)
	if err == nil {
		t.Fatal("expected error for context timeout, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
