package gateway

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	counters map[string]*rateCounter
	limit    int
	window   time.Duration
}

type rateCounter struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		counters: make(map[string]*rateCounter),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	counter, exists := rl.counters[key]

	if !exists || now.After(counter.resetAt) {
		rl.counters[key] = &rateCounter{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	if counter.count >= rl.limit {
		return false
	}

	counter.count++
	return true
}

func (rl *RateLimiter) Remaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	counter, exists := rl.counters[key]

	if !exists || now.After(counter.resetAt) {
		return rl.limit
	}

	remaining := rl.limit - counter.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (rl *RateLimiter) ResetAt(key string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	counter, exists := rl.counters[key]
	if !exists {
		return time.Time{}
	}
	return counter.resetAt
}

func (rl *RateLimiter) cleanup() {
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

func RateLimitMiddleware(limiter *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := GetPrincipalFromContext(r.Context())
		if principal == nil {
			next.ServeHTTP(w, r)
			return
		}

		key := principal.APIKeyID
		if !limiter.Allow(key) {
			resetAt := limiter.ResetAt(key)
			w.Header().Set("X-RateLimit-Limit", "60")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))
			w.Header().Set("Retry-After", "60")
			writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Rate limit exceeded. Please retry after some time.")
			return
		}

		remaining := limiter.Remaining(key)
		resetAt := limiter.ResetAt(key)
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", string(rune(remaining+'0')))
		if !resetAt.IsZero() {
			w.Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))
		}

		next.ServeHTTP(w, r)
	})
}
