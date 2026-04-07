package engine

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine/tool"
)

// ToolGateway provides a unified interface for executing various tools.
type ToolGateway struct {
	httpClient     *http.Client
	toolReg        *tool.Registry
	circuitBreaker *circuitBreaker
}

// GatewayToolResult represents the result of a tool invocation.
type GatewayToolResult struct {
	Status    string         `json:"status"`
	Output    map[string]any `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   time.Time      `json:"ended_at"`
}

// GatewayToolInvocation represents a tool invocation request.
type GatewayToolInvocation struct {
	ExecutionID    string         `json:"execution_id"`
	StepID         string         `json:"step_id"`
	ToolRef        string         `json:"tool_ref"`
	Input          map[string]any `json:"input"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	AttemptNo      int            `json:"attempt_no"`
}

// Circuit breaker state
type circuitBreakerState int

const (
	circuitBreakerClosed circuitBreakerState = iota
	circuitBreakerOpen
	circuitBreakerHalfOpen
)

type circuitBreaker struct {
	failureCount     int
	lastStateChange  time.Time
	state            circuitBreakerState
	failureThreshold int
	timeout          time.Duration
	mu               chan struct{}
}

// NewToolGateway creates a new ToolGateway with default HTTP client and no tool registry.
// For backward compatibility, this creates a mock gateway that doesn't execute real tools.
func NewToolGateway() *ToolGateway {
	return &ToolGateway{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		toolReg: nil,
		circuitBreaker: &circuitBreaker{
			failureThreshold: 5,
			timeout:          60 * time.Second,
			mu:               make(chan struct{}, 1),
		},
	}
}

// NewToolGatewayWithRegistry creates a new ToolGateway with the provided tool registry.
func NewToolGatewayWithRegistry(tr *tool.Registry) *ToolGateway {
	return &ToolGateway{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		toolReg: tr,
		circuitBreaker: &circuitBreaker{
			failureThreshold: 5,
			timeout:          60 * time.Second,
			mu:               make(chan struct{}, 1),
		},
	}
}

// Invoke executes a single tool invocation with circuit breaker protection.
func (gw *ToolGateway) Invoke(ctx context.Context, inv GatewayToolInvocation) (*GatewayToolResult, error) {
	start := time.Now()

	// Check if context is cancelled first
	select {
	case <-ctx.Done():
		return &GatewayToolResult{
			Status:    "failed",
			Error:     "context cancelled",
			StartedAt: start,
			EndedAt:   time.Now(),
		}, ctx.Err()
	default:
	}

	// Check circuit breaker state
	if !gw.circuitBreaker.allowRequest() {
		return &GatewayToolResult{
			Status:    "failed",
			Error:     "circuit breaker open",
			StartedAt: start,
			EndedAt:   time.Now(),
		}, fmt.Errorf("circuit breaker open")
	}

	// Find the tool by reference
	// If no registry is configured, return mock result (backward compatibility)
	if gw.toolReg == nil {
		select {
		case <-ctx.Done():
			return &GatewayToolResult{
				Status:    "failed",
				Error:     "context cancelled",
				StartedAt: start,
				EndedAt:   time.Now(),
			}, ctx.Err()
		default:
			return &GatewayToolResult{
				Status:    "succeeded",
				Output:    map[string]any{"mock": true, "tool_ref": inv.ToolRef, "execution_id": inv.ExecutionID},
				StartedAt: start,
				EndedAt:   time.Now(),
			}, nil
		}
	}

	t, exists := gw.toolReg.Get(inv.ToolRef)
	if !exists {
		gw.circuitBreaker.recordFailure()
		return &GatewayToolResult{
			Status:    "failed",
			Error:     fmt.Sprintf("tool not found: %s", inv.ToolRef),
			StartedAt: start,
			EndedAt:   time.Now(),
		}, fmt.Errorf("tool not found: %s", inv.ToolRef)
	}

	// Set up timeout context if specified
	var cancel context.CancelFunc
	if inv.TimeoutSeconds > 0 {
		var timeout time.Duration = time.Duration(inv.TimeoutSeconds) * time.Second
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Execute the tool
	result, err := t.Execute(ctx, inv.Input)
	if err != nil {
		gw.circuitBreaker.recordFailure()
		return &GatewayToolResult{
			Status:    "failed",
			Error:     err.Error(),
			StartedAt: start,
			EndedAt:   time.Now(),
		}, err
	}

	// Record success
	gw.circuitBreaker.recordSuccess()

	// Prepare output based on tool type
	output := map[string]any{
		"tool_ref":     inv.ToolRef,
		"execution_id": inv.ExecutionID,
		"step_id":      inv.StepID,
	}

	// If result is already a map, merge it; otherwise, put it under "result"
	if resultMap, ok := result.(map[string]any); ok {
		for k, v := range resultMap {
			output[k] = v
		}
	} else {
		output["result"] = result
	}

	return &GatewayToolResult{
		Status:    "succeeded",
		Output:    output,
		StartedAt: start,
		EndedAt:   time.Now(),
	}, nil
}

// InvokeWithRetry executes a tool invocation with retry logic and circuit breaker protection.
func (gw *ToolGateway) InvokeWithRetry(ctx context.Context, inv GatewayToolInvocation, maxAttempts int, backoffSeconds int) (*GatewayToolResult, error) {
	var lastResult *GatewayToolResult
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

// circuitBreaker methods
func (cb *circuitBreaker) allowRequest() bool {
	select {
	case cb.mu <- struct{}{}:
		defer func() { <-cb.mu }()

		switch cb.state {
		case circuitBreakerClosed:
			return true
		case circuitBreakerOpen:
			if time.Since(cb.lastStateChange) > cb.timeout {
				cb.state = circuitBreakerHalfOpen
				return true
			}
			return false
		case circuitBreakerHalfOpen:
			return true
		}
		return false
	}
}

func (cb *circuitBreaker) recordSuccess() {
	select {
	case cb.mu <- struct{}{}:
		defer func() { <-cb.mu }()
		cb.failureCount = 0
		cb.state = circuitBreakerClosed
	}
}

func (cb *circuitBreaker) recordFailure() {
	select {
	case cb.mu <- struct{}{}:
		defer func() { <-cb.mu }()
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = circuitBreakerOpen
			cb.lastStateChange = time.Now()
		}
	}
}
