package gateway

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting operations
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Remaining(ctx context.Context, key string) (int, error)
	ResetAt(ctx context.Context, key string) (time.Time, error)
	Reset(ctx context.Context, key string) error
}

// RateLimiterFactory creates rate limiters
type RateLimiterFactory interface {
	Create(keyPrefix string, limit int, window time.Duration) RateLimiter
}

// InMemoryRateLimiter is an in-memory implementation of RateLimiter
type InMemoryRateLimiter struct {
	mu       sync.RWMutex
	counters map[string]*rateCounter
	limit    int
	window   time.Duration
}

type rateCounter struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		counters: make(map[string]*rateCounter),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// NewRateLimiter creates a new in-memory rate limiter (for backward compatibility)
func NewRateLimiter(requestsPerMinute int) *InMemoryRateLimiter {
	return newRateLimiter(requestsPerMinute, time.Minute)
}

func (rl *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	counter, exists := rl.counters[key]

	if !exists || now.After(counter.resetAt) {
		rl.counters[key] = &rateCounter{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true, nil
	}

	if counter.count >= rl.limit {
		return false, nil
	}

	counter.count++
	return true, nil
}

func (rl *InMemoryRateLimiter) Remaining(ctx context.Context, key string) (int, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	counter, exists := rl.counters[key]

	if !exists || now.After(counter.resetAt) {
		return rl.limit, nil
	}

	remaining := rl.limit - counter.count
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (rl *InMemoryRateLimiter) ResetAt(ctx context.Context, key string) (time.Time, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	counter, exists := rl.counters[key]
	if !exists {
		return time.Time{}, nil
	}
	return counter.resetAt, nil
}

func (rl *InMemoryRateLimiter) Reset(ctx context.Context, key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.counters, key)
	return nil
}

func (rl *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, counter := range rl.counters {
			if now.After(counter.resetAt) {
				delete(rl.counters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// InMemoryRateLimiterFactory creates in-memory rate limiters
type InMemoryRateLimiterFactory struct{}

func (f *InMemoryRateLimiterFactory) Create(keyPrefix string, limit int, window time.Duration) RateLimiter {
	return newRateLimiter(limit, window)
}

func RateLimitMiddleware(limiter RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := GetPrincipalFromContext(r.Context())
		if principal == nil {
			next.ServeHTTP(w, r)
			return
		}

		key := principal.APIKeyID
		allowed, err := limiter.Allow(r.Context(), key)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "rate_limit_error", "Rate limit check failed.")
			return
		}
		if !allowed {
			resetAt, _ := limiter.ResetAt(r.Context(), key)
			w.Header().Set("X-RateLimit-Limit", "60")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))
			w.Header().Set("Retry-After", "60")
			writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Rate limit exceeded. Please retry after some time.")
			return
		}

		remaining, _ := limiter.Remaining(r.Context(), key)
		resetAt, _ := limiter.ResetAt(r.Context(), key)
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", string(rune(remaining+'0')))
		if !resetAt.IsZero() {
			w.Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))
		}

		next.ServeHTTP(w, r)
	})
}
