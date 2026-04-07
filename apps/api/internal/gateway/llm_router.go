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

// LLMRouter routes requests to LLM providers with fallback support
type LLMRouter interface {
	Route(ctx context.Context) (llm.Client, error)
	RegisterProvider(config ProviderConfig)
	RecordFailure(provider llm.Provider)
	RecordSuccess(provider llm.Provider)
	GetCurrentProvider() llm.Provider
}

// FallbackRouter implements LLMRouter with priority-based fallback
type FallbackRouter struct {
	mu             sync.RWMutex
	providers      []ProviderConfig
	circuitBreaker CircuitBreaker
	currentIndex   int
}

// NewFallbackRouter creates a new FallbackRouter with the given circuit breaker
func NewFallbackRouter(cb CircuitBreaker) *FallbackRouter {
	return &FallbackRouter{
		providers:      make([]ProviderConfig, 0),
		circuitBreaker: cb,
		currentIndex:   0,
	}
}

// RegisterProvider adds a provider to the router
func (r *FallbackRouter) RegisterProvider(config ProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Insert in priority order
	r.providers = append(r.providers, config)
	sort.Slice(r.providers, func(i, j int) bool {
		return r.providers[i].Priority < r.providers[j].Priority
	})
}

// Route selects the first available provider based on priority and circuit breaker state
func (r *FallbackRouter) Route(ctx context.Context) (llm.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}

	for i, p := range r.providers {
		key := string(p.Provider)
		allowed, err := r.circuitBreaker.Allow(ctx, key)
		if err == nil && allowed {
			r.currentIndex = i
			return p.Client, nil
		}
	}

	return nil, fmt.Errorf("all providers unavailable (circuit breakers open)")
}

// RecordFailure records a failure for the provider and updates circuit breaker
func (r *FallbackRouter) RecordFailure(provider llm.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := string(provider)
	_ = r.circuitBreaker.RecordFailure(context.Background(), key)

	if r.currentIndex < len(r.providers)-1 {
		r.currentIndex++
	}
}

// RecordSuccess records a success for the provider and updates circuit breaker
func (r *FallbackRouter) RecordSuccess(provider llm.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := string(provider)
	_ = r.circuitBreaker.RecordSuccess(context.Background(), key)
}

// GetCurrentProvider returns the currently selected provider
func (r *FallbackRouter) GetCurrentProvider() llm.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.currentIndex < len(r.providers) {
		return r.providers[r.currentIndex].Provider
	}
	return ""
}
