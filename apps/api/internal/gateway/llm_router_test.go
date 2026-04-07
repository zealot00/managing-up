package gateway

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type mockLLMClient struct {
	providerValue llm.Provider
	modelValue    llm.Model
	generateFunc  func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error)
	failCount     atomic.Int32
}

func (m *mockLLMClient) Generate(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, messages, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLLMClient) GenerateStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (llm.StreamReader, error) {
	return nil, errors.New("not implemented")
}

func (m *mockLLMClient) Provider() llm.Provider {
	return m.providerValue
}

func (m *mockLLMClient) Model() llm.Model {
	return m.modelValue
}

type mockCircuitBreaker struct {
	allowFunc         func(ctx context.Context, key string) (bool, error)
	recordSuccessFunc func(ctx context.Context, key string) error
	recordFailureFunc func(ctx context.Context, key string) error
	states            map[string]CircuitBreakerState
}

func newMockCircuitBreaker() *mockCircuitBreaker {
	return &mockCircuitBreaker{
		states: make(map[string]CircuitBreakerState),
	}
}

func (m *mockCircuitBreaker) Allow(ctx context.Context, key string) (bool, error) {
	if m.allowFunc != nil {
		return m.allowFunc(ctx, key)
	}
	if state, ok := m.states[key]; ok {
		return state != CircuitBreakerOpen, nil
	}
	return true, nil
}

func (m *mockCircuitBreaker) RecordSuccess(ctx context.Context, key string) error {
	if m.recordSuccessFunc != nil {
		return m.recordSuccessFunc(ctx, key)
	}
	if state, ok := m.states[key]; ok && state == CircuitBreakerHalfOpen {
		m.states[key] = CircuitBreakerClosed
	}
	return nil
}

func (m *mockCircuitBreaker) RecordFailure(ctx context.Context, key string) error {
	if m.recordFailureFunc != nil {
		return m.recordFailureFunc(ctx, key)
	}
	if _, ok := m.states[key]; !ok {
		m.states[key] = CircuitBreakerClosed
	}
	m.states[key] = CircuitBreakerOpen
	return nil
}

func (m *mockCircuitBreaker) Reset(ctx context.Context, key string) error {
	delete(m.states, key)
	return nil
}

func (m *mockCircuitBreaker) State(ctx context.Context, key string) (CircuitBreakerState, error) {
	if state, ok := m.states[key]; ok {
		return state, nil
	}
	return CircuitBreakerClosed, nil
}

func TestFallbackRouter_BasicFallback(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	failCalled := atomic.Bool{}
	primaryCalled := atomic.Bool{}

	primaryClient := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			primaryCalled.Store(true)
			return nil, errors.New("primary failed")
		},
	}

	fallbackClient := &mockLLMClient{
		providerValue: "anthropic",
		modelValue:    "claude-sonnet-4",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			failCalled.Store(true)
			return &llm.Response{
				Content:  "fallback success",
				Model:    "claude-sonnet-4",
				Provider: "anthropic",
			}, nil
		},
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   primaryClient,
		Priority: 1,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "anthropic",
		Client:   fallbackClient,
		Priority: 2,
	})

	ctx := context.Background()
	resp, err := GenerateWithRouterRetry(ctx, router, []llm.Message{{Role: "user", Content: "hello"}}, nil, RetryConfig{
		MaxRetries:     1,
		InitialBackoff: 1,
		MaxBackoff:     1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Content != "fallback success" {
		t.Errorf("expected 'fallback success', got %s", resp.Content)
	}
	if !failCalled.Load() {
		t.Error("expected fallback to be called")
	}
}

func TestFallbackRouter_AllProvidersFail(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	primaryClient := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			return nil, errors.New("primary failed")
		},
	}

	fallbackClient := &mockLLMClient{
		providerValue: "anthropic",
		modelValue:    "claude-sonnet-4",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			return nil, errors.New("fallback failed")
		},
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   primaryClient,
		Priority: 1,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "anthropic",
		Client:   fallbackClient,
		Priority: 2,
	})

	ctx := context.Background()
	_, err := GenerateWithRouterRetry(ctx, router, []llm.Message{{Role: "user", Content: "hello"}}, nil, RetryConfig{
		MaxRetries:     0,
		InitialBackoff: 1,
		MaxBackoff:     1,
	})

	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestFallbackRouter_CircuitBreakerIntegration(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	fallbackCalled := atomic.Bool{}

	primaryClient := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			return nil, errors.New("primary circuit broken")
		},
	}

	fallbackClient := &mockLLMClient{
		providerValue: "anthropic",
		modelValue:    "claude-sonnet-4",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			fallbackCalled.Store(true)
			return &llm.Response{
				Content:  "fallback success",
				Model:    "claude-sonnet-4",
				Provider: "anthropic",
			}, nil
		},
	}

	cb.states["openai"] = CircuitBreakerOpen

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   primaryClient,
		Priority: 1,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "anthropic",
		Client:   fallbackClient,
		Priority: 2,
	})

	ctx := context.Background()
	resp, err := GenerateWithRouterRetry(ctx, router, []llm.Message{{Role: "user", Content: "hello"}}, nil, RetryConfig{
		MaxRetries:     0,
		InitialBackoff: 1,
		MaxBackoff:     1,
	})

	if err != nil {
		t.Fatalf("expected no error with circuit breaker open on primary, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response from fallback")
	}
	if resp.Content != "fallback success" {
		t.Errorf("expected 'fallback success', got %s", resp.Content)
	}
	if !fallbackCalled.Load() {
		t.Error("expected fallback to be called when primary circuit is open")
	}
}

func TestFallbackRouter_PriorityOrder(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	callOrder := atomic.Int32{}

	firstClient := &mockLLMClient{
		providerValue: "first",
		modelValue:    "model-1",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			return &llm.Response{
				Content:  "first success",
				Model:    "model-1",
				Provider: "first",
			}, nil
		},
	}

	secondClient := &mockLLMClient{
		providerValue: "second",
		modelValue:    "model-2",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			callOrder.Add(1)
			return nil, errors.New("second failed")
		},
	}

	thirdClient := &mockLLMClient{
		providerValue: "third",
		modelValue:    "model-3",
		generateFunc: func(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
			callOrder.Add(10)
			return nil, errors.New("third failed")
		},
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "third",
		Client:   thirdClient,
		Priority: 3,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "second",
		Client:   secondClient,
		Priority: 2,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "first",
		Client:   firstClient,
		Priority: 1,
	})

	ctx := context.Background()
	resp, err := GenerateWithRouterRetry(ctx, router, []llm.Message{{Role: "user", Content: "hello"}}, nil, RetryConfig{
		MaxRetries:     0,
		InitialBackoff: 1,
		MaxBackoff:     1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Content != "first success" {
		t.Errorf("expected 'first success' (highest priority), got %s", resp.Content)
	}
	if callOrder.Load() != 0 {
		t.Error("lower priority providers should not be called when higher priority succeeds")
	}
}

func TestFallbackRouter_GetCurrentProvider(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	client1 := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
	}
	client2 := &mockLLMClient{
		providerValue: "anthropic",
		modelValue:    "claude-sonnet-4",
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   client1,
		Priority: 1,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "anthropic",
		Client:   client2,
		Priority: 2,
	})

	ctx := context.Background()
	_, err := router.Route(ctx)
	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	provider := router.GetCurrentProvider()
	if provider != "openai" {
		t.Errorf("expected current provider 'openai', got %s", provider)
	}
}

func TestFallbackRouter_RecordFailureUpdatesCircuitBreaker(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	client := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   client,
		Priority: 1,
	})

	ctx := context.Background()
	router.RecordFailure("openai")

	state, err := cb.State(ctx, "openai")
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerOpen {
		t.Errorf("expected circuit breaker to be open after failure, got %s", state)
	}

	cb.states["openai"] = CircuitBreakerHalfOpen
	router.RecordSuccess("openai")

	state, err = cb.State(ctx, "openai")
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected circuit breaker to be closed after success in half-open state, got %s", state)
	}
}

func TestFallbackRouter_RecordFailureAdvancesIndex(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	client1 := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
	}
	client2 := &mockLLMClient{
		providerValue: "anthropic",
		modelValue:    "claude-sonnet-4",
	}
	client3 := &mockLLMClient{
		providerValue: "google",
		modelValue:    "gemini-2.0-flash",
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   client1,
		Priority: 1,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "anthropic",
		Client:   client2,
		Priority: 2,
	})
	router.RegisterProvider(ProviderConfig{
		Provider: "google",
		Client:   client3,
		Priority: 3,
	})

	initialProvider := router.GetCurrentProvider()
	if initialProvider != "openai" {
		t.Fatalf("expected initial provider 'openai', got %s", initialProvider)
	}

	router.RecordFailure("openai")
	afterFirstFailure := router.GetCurrentProvider()
	if afterFirstFailure != "anthropic" {
		t.Errorf("expected provider to advance to 'anthropic' after first failure, got %s", afterFirstFailure)
	}

	router.RecordFailure("anthropic")
	afterSecondFailure := router.GetCurrentProvider()
	if afterSecondFailure != "google" {
		t.Errorf("expected provider to advance to 'google' after second failure, got %s", afterSecondFailure)
	}

	router.RecordFailure("google")
	afterThirdFailure := router.GetCurrentProvider()
	if afterThirdFailure != "google" {
		t.Errorf("expected provider to stay at 'google' (last in list) after third failure, got %s", afterThirdFailure)
	}
}

func TestFallbackRouter_RecordFailureDoesNotAdvanceBeyondLast(t *testing.T) {
	cb := newMockCircuitBreaker()
	router := NewFallbackRouter(cb)

	client := &mockLLMClient{
		providerValue: "openai",
		modelValue:    "gpt-4o",
	}

	router.RegisterProvider(ProviderConfig{
		Provider: "openai",
		Client:   client,
		Priority: 1,
	})

	for i := 0; i < 5; i++ {
		router.RecordFailure("openai")
	}

	provider := router.GetCurrentProvider()
	if provider != "openai" {
		t.Errorf("expected provider to stay at 'openai' (only provider), got %s", provider)
	}

	state, _ := cb.State(context.Background(), "openai")
	if state != CircuitBreakerOpen {
		t.Errorf("expected circuit breaker to be open after failures, got %s", state)
	}
}
