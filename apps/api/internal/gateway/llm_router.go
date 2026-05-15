package gateway

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// ProviderConfig holds configuration for an LLM provider
type ProviderConfig struct {
	Provider llm.Provider
	Client   llm.Client
	Priority int // Lower = preferred
}

// FallbackTarget represents a single model in a fallback chain.
type FallbackTarget struct {
	Provider llm.Provider
	Model    llm.Model
}

// LLMRouter routes requests to LLM providers with fallback support
type LLMRouter interface {
	Route(ctx context.Context) (llm.Client, error)
	RegisterProvider(config ProviderConfig)
	RecordFailure(provider llm.Provider)
	RecordSuccess(provider llm.Provider)
	GetCurrentProvider() llm.Provider
	RouteWithFallback(ctx context.Context, model llm.Model, resolveKey func(llm.Provider) string) ([]llm.Client, error)
}

// FallbackRouter implements LLMRouter with priority-based and model-level fallback
type FallbackRouter struct {
	mu             sync.RWMutex
	providers      []ProviderConfig
	fallbackChains map[llm.Model][]FallbackTarget
	providerKeys   map[llm.Provider]string
	circuitBreaker CircuitBreaker
}

// NewFallbackRouter creates a new FallbackRouter with the given circuit breaker
func NewFallbackRouter(cb CircuitBreaker) *FallbackRouter {
	return &FallbackRouter{
		providers:      make([]ProviderConfig, 0),
		fallbackChains: make(map[llm.Model][]FallbackTarget),
		providerKeys:   make(map[llm.Provider]string),
		circuitBreaker: cb,
	}
}

// RegisterProvider adds a provider to the router
func (r *FallbackRouter) RegisterProvider(config ProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = append(r.providers, config)
	sort.Slice(r.providers, func(i, j int) bool {
		return r.providers[i].Priority < r.providers[j].Priority
	})
}

// SetFallbackChains configures per-model fallback chains.
func (r *FallbackRouter) SetFallbackChains(chains map[llm.Model][]FallbackTarget) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallbackChains = chains
}

// SetProviderKeys configures API keys for creating clients during fallback.
func (r *FallbackRouter) SetProviderKeys(keys map[llm.Provider]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providerKeys = keys
}

// Route selects the first available provider based on priority and circuit breaker state
func (r *FallbackRouter) Route(ctx context.Context) (llm.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}

	for _, p := range r.providers {
		key := string(p.Provider)
		allowed, err := r.circuitBreaker.Allow(ctx, key)
		if err == nil && allowed {
			return p.Client, nil
		}
	}

	return nil, fmt.Errorf("all providers unavailable (circuit breakers open)")
}

// RouteWithFallback returns a list of clients ordered by fallback priority
// for the given model. The first client is the preferred provider;
// subsequent clients are fallback targets.
// resolveKey is called to get the API key for a provider — it may return ""
// to use the router's built-in providerKeys.
func (r *FallbackRouter) RouteWithFallback(ctx context.Context, model llm.Model, resolveKey func(llm.Provider) string) ([]llm.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chain, hasChain := r.fallbackChains[model]

	if !hasChain {
		// No fallback chain configured — try to create a single client
		// using the model's inferred provider
		provider, _, err := llm.ParseModelString(string(model))
		if err != nil {
			return nil, fmt.Errorf("cannot determine provider for model %s: %w", model, err)
		}
		client, err := r.createClient(provider, model, resolveKey)
		if err != nil {
			return nil, err
		}
		return []llm.Client{client}, nil
	}

	var clients []llm.Client
	for _, target := range chain {
		// Check circuit breaker for this provider
		key := string(target.Provider)
		allowed, err := r.circuitBreaker.Allow(ctx, key)
		if err != nil || !allowed {
			continue
		}

		client, err := r.createClient(target.Provider, target.Model, resolveKey)
		if err != nil {
			continue
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("all providers unavailable for model %s (circuit breakers open or no API keys)", model)
	}

	return clients, nil
}

// createClient creates an LLM client for the given provider and model.
func (r *FallbackRouter) createClient(provider llm.Provider, model llm.Model, resolveKey func(llm.Provider) string) (llm.Client, error) {
	apiKey := ""
	if resolveKey != nil {
		apiKey = resolveKey(provider)
	}
	if apiKey == "" {
		apiKey = r.providerKeys[provider]
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no API key configured for provider %s", provider)
	}
	return llm.NewClient(provider, model, apiKey)
}

// RecordFailure records a failure for the provider and updates circuit breaker
func (r *FallbackRouter) RecordFailure(provider llm.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := string(provider)
	_ = r.circuitBreaker.RecordFailure(context.Background(), key)
}

// RecordSuccess records a success for the provider and updates circuit breaker
func (r *FallbackRouter) RecordSuccess(provider llm.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := string(provider)
	_ = r.circuitBreaker.RecordSuccess(context.Background(), key)
}

func (r *FallbackRouter) GetCurrentProvider() llm.Provider {
	return ""
}

// GetFallbackChains returns a copy of the current fallback chains (for inspection/testing).
func (r *FallbackRouter) GetFallbackChains() map[llm.Model][]FallbackTarget {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[llm.Model][]FallbackTarget, len(r.fallbackChains))
	for k, v := range r.fallbackChains {
		cp := make([]FallbackTarget, len(v))
		copy(cp, v)
		result[k] = cp
	}
	return result
}
