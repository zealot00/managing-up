package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// --- Mocks ---

type mockCircuitBreaker struct {
	allowed map[string]bool
}

func (m *mockCircuitBreaker) Allow(ctx context.Context, key string) (bool, error) {
	allowed, ok := m.allowed[key]
	if !ok {
		return true, nil // default: allow
	}
	return allowed, nil
}

func (m *mockCircuitBreaker) RecordSuccess(ctx context.Context, key string) error { return nil }
func (m *mockCircuitBreaker) RecordFailure(ctx context.Context, key string) error { return nil }
func (m *mockCircuitBreaker) Reset(ctx context.Context, key string) error          { return nil }

type mockLLMClient struct {
	provider llm.Provider
	model    llm.Model
}

func (m *mockLLMClient) Generate(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
	return nil, nil
}
func (m *mockLLMClient) GenerateStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (llm.StreamReader, error) {
	return nil, nil
}
func (m *mockLLMClient) Provider() llm.Provider { return m.provider }
func (m *mockLLMClient) Model() llm.Model        { return m.model }

// --- Tests ---

func TestNewFallbackRouter(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestFallbackRouter_RegisterProvider(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)

	r.RegisterProvider(ProviderConfig{
		Provider: llm.ProviderOpenAI,
		Client:   &mockLLMClient{provider: llm.ProviderOpenAI, model: "gpt-4o"},
		Priority: 2,
	})
	r.RegisterProvider(ProviderConfig{
		Provider: llm.ProviderAnthropic,
		Client:   &mockLLMClient{provider: llm.ProviderAnthropic, model: "claude-sonnet-4"},
		Priority: 1,
	})

	// Route should return Anthropic (lower priority number = preferred)
	client, err := r.Route(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Provider() != llm.ProviderAnthropic {
		t.Errorf("expected Anthropic (priority 1), got %s", client.Provider())
	}
}

func TestFallbackRouter_Route_CircuitBreakerOpen(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{allowed: map[string]bool{
		"openai": false,
	}}

	r := NewFallbackRouter(cb)
	r.RegisterProvider(ProviderConfig{
		Provider: llm.ProviderOpenAI,
		Client:   &mockLLMClient{provider: llm.ProviderOpenAI},
		Priority: 1,
	})

	_, err := r.Route(context.Background())
	if err == nil {
		t.Fatal("expected error when all circuit breakers are open")
	}
}

func TestFallbackRouter_RouteWithFallback_NoChain(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	r.SetProviderKeys(map[llm.Provider]string{
		llm.ProviderOpenAI: "test-key",
	})

	// No fallback chain configured for this model
	clients, err := r.RouteWithFallback(context.Background(), "gpt-4o", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if clients[0].Provider() != llm.ProviderOpenAI {
		t.Errorf("expected OpenAI, got %s", clients[0].Provider())
	}
}

func TestFallbackRouter_RouteWithFallback_WithChain(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	r.SetProviderKeys(map[llm.Provider]string{
		llm.ProviderOpenAI:    "openai-key",
		llm.ProviderAnthropic: "anthropic-key",
		llm.ProviderOllama:    "http://localhost:11434",
	})
	r.SetFallbackChains(map[llm.Model][]FallbackTarget{
		"gpt-4o": {
			{Provider: llm.ProviderOpenAI, Model: "gpt-4o"},
			{Provider: llm.ProviderAnthropic, Model: "claude-sonnet-4"},
			{Provider: llm.ProviderOllama, Model: "qwen2.5"},
		},
	})

	clients, err := r.RouteWithFallback(context.Background(), "gpt-4o", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 3 {
		t.Fatalf("expected 3 clients, got %d", len(clients))
	}
	if clients[0].Provider() != llm.ProviderOpenAI {
		t.Errorf("expected first client to be OpenAI, got %s", clients[0].Provider())
	}
	if clients[1].Provider() != llm.ProviderAnthropic {
		t.Errorf("expected second client to be Anthropic, got %s", clients[1].Provider())
	}
}

func TestFallbackRouter_RouteWithFallback_CircuitBreakerSkipsProvider(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{allowed: map[string]bool{
		"openai": false, // OpenAI circuit breaker is open
	}}
	r := NewFallbackRouter(cb)
	r.SetProviderKeys(map[llm.Provider]string{
		llm.ProviderOpenAI:    "openai-key",
		llm.ProviderAnthropic: "anthropic-key",
	})
	r.SetFallbackChains(map[llm.Model][]FallbackTarget{
		"gpt-4o": {
			{Provider: llm.ProviderOpenAI, Model: "gpt-4o"},
			{Provider: llm.ProviderAnthropic, Model: "claude-sonnet-4"},
		},
	})

	clients, err := r.RouteWithFallback(context.Background(), "gpt-4o", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client (OpenAI skipped by CB), got %d", len(clients))
	}
	if clients[0].Provider() != llm.ProviderAnthropic {
		t.Errorf("expected Anthropic, got %s", clients[0].Provider())
	}
}

func TestFallbackRouter_RouteWithFallback_ResolveKey(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	// No provider keys in router — resolveKey must provide them

	resolveKey := func(p llm.Provider) string {
		if p == llm.ProviderOpenAI {
			return "user-provided-openai-key"
		}
		return ""
	}

	clients, err := r.RouteWithFallback(context.Background(), "gpt-4o", resolveKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
}

func TestFallbackRouter_RouteWithFallback_AllCBsOpen(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{allowed: map[string]bool{
		"openai":    false,
		"anthropic": false,
	}}
	r := NewFallbackRouter(cb)
	r.SetProviderKeys(map[llm.Provider]string{
		llm.ProviderOpenAI:    "key",
		llm.ProviderAnthropic: "key",
	})
	r.SetFallbackChains(map[llm.Model][]FallbackTarget{
		"gpt-4o": {
			{Provider: llm.ProviderOpenAI, Model: "gpt-4o"},
			{Provider: llm.ProviderAnthropic, Model: "claude-sonnet-4"},
		},
	})

	_, err := r.RouteWithFallback(context.Background(), "gpt-4o", nil)
	if err == nil {
		t.Fatal("expected error when all circuit breakers are open")
	}
}

func TestFallbackRouter_RouteWithFallback_NoAPIKey(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	// No provider keys and no resolveKey

	_, err := r.RouteWithFallback(context.Background(), "gpt-4o", nil)
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
}

func TestFallbackRouter_GetFallbackChains(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	r.SetFallbackChains(map[llm.Model][]FallbackTarget{
		"gpt-4o": {
			{Provider: llm.ProviderAnthropic, Model: "claude-sonnet-4"},
		},
	})

	chains := r.GetFallbackChains()
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	targets, ok := chains["gpt-4o"]
	if !ok {
		t.Fatal("expected gpt-4o chain")
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	// Verify it's a copy (modifying returned map doesn't affect router)
	chains["test"] = []FallbackTarget{}
	if _, ok := r.GetFallbackChains()["test"]; ok {
		t.Error("GetFallbackChains should return a copy")
	}
}

func TestFallbackRouter_RouteWithFallback_ResolveKeyOverridesProviderKeys(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{}
	r := NewFallbackRouter(cb)
	r.SetProviderKeys(map[llm.Provider]string{
		llm.ProviderOpenAI: "default-key",
	})

	resolveKey := func(p llm.Provider) string {
		return "override-key"
	}

	clients, err := r.RouteWithFallback(context.Background(), "gpt-4o", resolveKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
}

func TestClassifyError_Integration(t *testing.T) {
	t.Parallel()

	// Verify that ProviderError from llm package integrates correctly
	err := &llm.ProviderError{
		Provider:   llm.ProviderOpenAI,
		StatusCode: 429,
		Message:    "rate limited",
	}

	class := llm.ClassifyError(err)
	if class != llm.ErrorClassFallbackable {
		t.Errorf("expected Fallbackable for 429, got %v", class)
	}

	// Non-retryable
	err2 := &llm.ProviderError{
		Provider:   llm.ProviderAnthropic,
		StatusCode: 401,
		Message:    "unauthorized",
	}
	class2 := llm.ClassifyError(err2)
	if class2 != llm.ErrorClassNonRetryable {
		t.Errorf("expected NonRetryable for 401, got %v", class2)
	}

	// Legacy string error
	err3 := errors.New("OpenAI API returned status 503: service unavailable")
	class3 := llm.ClassifyError(err3)
	if class3 != llm.ErrorClassFallbackable {
		t.Errorf("expected Fallbackable for 503 string, got %v", class3)
	}
}
