package engine

import (
	"context"
	"net/http"
	"time"
)

type ToolGateway struct {
	httpClient *http.Client
}

type ToolResult struct {
	Status    string         `json:"status"`
	Output    map[string]any `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   time.Time      `json:"ended_at"`
}

type ToolInvocation struct {
	ExecutionID    string         `json:"execution_id"`
	StepID         string         `json:"step_id"`
	ToolRef        string         `json:"tool_ref"`
	Input          map[string]any `json:"input"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	AttemptNo      int            `json:"attempt_no"`
}

func NewToolGateway() *ToolGateway {
	return &ToolGateway{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (gw *ToolGateway) Invoke(ctx context.Context, inv ToolInvocation) (*ToolResult, error) {
	start := time.Now()

	select {
	case <-ctx.Done():
		return &ToolResult{
			Status:    "failed",
			Error:     "context cancelled",
			StartedAt: start,
			EndedAt:   time.Now(),
		}, ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	return &ToolResult{
		Status: "succeeded",
		Output: map[string]any{
			"mock":         true,
			"tool_ref":     inv.ToolRef,
			"execution_id": inv.ExecutionID,
			"step_id":      inv.StepID,
		},
		StartedAt: start,
		EndedAt:   time.Now(),
	}, nil
}

func (gw *ToolGateway) InvokeWithRetry(ctx context.Context, inv ToolInvocation, maxAttempts int, backoffSeconds int) (*ToolResult, error) {
	var lastResult *ToolResult
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		inv.AttemptNo = attempt
		result, err := gw.Invoke(ctx, inv)
		if err == nil && result.Status == "succeeded" {
			return result, nil
		}

		lastResult = result
		lastErr = err

		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(time.Duration(backoffSeconds) * time.Second):
			}
		}
	}

	return lastResult, lastErr
}
