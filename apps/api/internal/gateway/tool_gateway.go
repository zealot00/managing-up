package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CircuitBreaker interface defines the contract for circuit breaker implementations
type CircuitBreaker interface {
	Allow(ctx context.Context, key string) (bool, error)
	RecordSuccess(ctx context.Context, key string) error
	RecordFailure(ctx context.Context, key string) error
	Reset(ctx context.Context, key string) error
}

// inMemoryCircuitBreaker implements CircuitBreaker in memory
type inMemoryCircuitBreaker struct {
	mu               sync.RWMutex
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	states           map[string]*circuitState
}

type circuitState struct {
	state           CircuitBreakerState
	failures        int
	successes       int
	lastFailureTime time.Time
	lastStateChange time.Time
}

// NewInMemoryCircuitBreaker creates an in-memory circuit breaker
func NewInMemoryCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *inMemoryCircuitBreaker {
	return &inMemoryCircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		states:           make(map[string]*circuitState),
	}
}

func (cb *inMemoryCircuitBreaker) Allow(ctx context.Context, key string) (bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.states[key]
	now := time.Now()

	if !exists {
		cb.states[key] = &circuitState{
			state:           CircuitBreakerClosed,
			lastStateChange: now,
		}
		return true, nil
	}

	switch state.state {
	case CircuitBreakerClosed:
		return true, nil
	case CircuitBreakerOpen:
		if now.Sub(state.lastFailureTime) >= cb.timeout {
			state.state = CircuitBreakerHalfOpen
			state.lastStateChange = now
			state.successes = 0
			return true, nil
		}
		return false, nil
	case CircuitBreakerHalfOpen:
		return true, nil
	}
	return false, nil
}

func (cb *inMemoryCircuitBreaker) RecordSuccess(ctx context.Context, key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.states[key]
	if !exists {
		return nil
	}

	if state.state == CircuitBreakerHalfOpen {
		state.successes++
		if state.successes >= cb.successThreshold {
			state.state = CircuitBreakerClosed
			state.lastStateChange = time.Now()
			state.failures = 0
			state.successes = 0
		}
	}
	return nil
}

func (cb *inMemoryCircuitBreaker) RecordFailure(ctx context.Context, key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.states[key]
	if !exists {
		cb.states[key] = &circuitState{
			state:           CircuitBreakerClosed,
			failures:        1,
			lastFailureTime: time.Now(),
		}
		return nil
	}

	state.lastFailureTime = time.Now()

	if state.state == CircuitBreakerHalfOpen {
		state.state = CircuitBreakerOpen
		state.lastStateChange = time.Now()
		return nil
	}

	if state.state == CircuitBreakerClosed {
		state.failures++
		if state.failures >= cb.failureThreshold {
			state.state = CircuitBreakerOpen
			state.lastStateChange = time.Now()
		}
	}
	return nil
}

func (cb *inMemoryCircuitBreaker) Reset(ctx context.Context, key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.states, key)
	return nil
}

func (cb *inMemoryCircuitBreaker) State(ctx context.Context, key string) (CircuitBreakerState, error) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state, exists := cb.states[key]
	if !exists {
		return CircuitBreakerClosed, nil
	}
	return state.state, nil
}

// ToolGateway handles tool invocations with circuit breaker support
type ToolGateway struct {
	httpClient     *http.Client
	circuitBreaker CircuitBreaker
}

// NewToolGateway creates a ToolGateway with in-memory circuit breaker
func NewToolGateway() *ToolGateway {
	return &ToolGateway{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		circuitBreaker: &inMemoryCircuitBreaker{
			failureThreshold: 5,
			successThreshold: 2,
			timeout:          60 * time.Second,
		},
	}
}

// NewToolGatewayWithCircuitBreaker creates a ToolGateway with a custom circuit breaker
func NewToolGatewayWithCircuitBreaker(cb CircuitBreaker) *ToolGateway {
	return &ToolGateway{
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		circuitBreaker: cb,
	}
}

// Invoke executes a tool invocation with circuit breaker protection
func (gw *ToolGateway) Invoke(ctx context.Context, toolRef string) error {
	allowed, err := gw.circuitBreaker.Allow(ctx, toolRef)
	if err != nil {
		return err
	}
	if !allowed {
		gw.circuitBreaker.RecordFailure(ctx, toolRef)
		return fmt.Errorf("circuit breaker open for tool: %s", toolRef)
	}

	// Simulate tool execution - in real implementation, would call actual tool
	select {
	case <-ctx.Done():
		gw.circuitBreaker.RecordFailure(ctx, toolRef)
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	gw.circuitBreaker.RecordSuccess(ctx, toolRef)
	return nil
}
