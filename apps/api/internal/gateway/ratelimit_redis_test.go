package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiter_Allow(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	limiter := NewRedisRateLimiter(client, "test:allow", 3, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Errorf("4th request should be blocked")
	}

	limiter.Reset(ctx, "key1")

	allowed, err = limiter.Allow(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if !allowed {
		t.Errorf("request after reset should be allowed")
	}
}

func TestRedisRateLimiter_Remaining(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	limiter := NewRedisRateLimiter(client, "test:remaining", 5, time.Minute)
	ctx := context.Background()

	remaining, err := limiter.Remaining(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 5 {
		t.Errorf("expected initial remaining 5, got %d", remaining)
	}

	limiter.Allow(ctx, "key1")
	remaining, err = limiter.Remaining(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 4 {
		t.Errorf("expected remaining 4 after 1 request, got %d", remaining)
	}

	limiter.Allow(ctx, "key1")
	limiter.Allow(ctx, "key1")
	remaining, err = limiter.Remaining(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 2 {
		t.Errorf("expected remaining 2 after 3 requests, got %d", remaining)
	}
}

func TestRedisRateLimiter_ResetAt(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	limiter := NewRedisRateLimiter(client, "test:resetat", 5, time.Minute)
	ctx := context.Background()

	resetAt, err := limiter.ResetAt(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resetAt.IsZero() {
		t.Errorf("expected zero time for unknown key, got %v", resetAt)
	}

	limiter.Allow(ctx, "key1")

	resetAt, err = limiter.ResetAt(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resetAt.IsZero() {
		t.Errorf("expected non-zero time after request")
	}

	expectedWindow := time.Now().Add(time.Minute)
	if resetAt.Before(expectedWindow.Add(-time.Second)) || resetAt.After(expectedWindow.Add(time.Second)) {
		t.Errorf("resetAt %v is not within expected window %v", resetAt, expectedWindow)
	}
}

func TestRedisRateLimiter_Reset(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	limiter := NewRedisRateLimiter(client, "test:reset", 5, time.Minute)
	ctx := context.Background()

	limiter.Allow(ctx, "key1")

	remaining, _ := limiter.Remaining(ctx, "key1")
	if remaining != 4 {
		t.Errorf("expected remaining 4, got %d", remaining)
	}

	err := limiter.Reset(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remaining, _ = limiter.Remaining(ctx, "key1")
	if remaining != 5 {
		t.Errorf("expected remaining 5 after reset, got %d", remaining)
	}
}

func TestRedisRateLimiterFactory_Create(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	factory := &RedisRateLimiterFactory{Client: client}
	limiter := factory.Create("test:factory", 10, time.Minute)

	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}

	ctx := context.Background()
	allowed, err := limiter.Allow(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected first request to be allowed")
	}
}

func TestRedisRateLimiter_MultipleKeys(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	limiter := NewRedisRateLimiter(client, "test:keys", 2, time.Minute)
	ctx := context.Background()

	limiter.Allow(ctx, "key1")
	limiter.Allow(ctx, "key1")
	allowed, _ := limiter.Allow(ctx, "key1")
	if allowed {
		t.Error("key1 should be rate limited")
	}

	limiter.Allow(ctx, "key2")
	limiter.Allow(ctx, "key2")
	allowed, _ = limiter.Allow(ctx, "key2")
	if allowed {
		t.Error("key2 should be rate limited")
	}

	allowed, _ = limiter.Allow(ctx, "key3")
	if !allowed {
		t.Error("key3 should still be allowed")
	}
}

func TestRateLimitMiddleware_Redis(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	factory := &RedisRateLimiterFactory{Client: client}
	limiter := factory.Create("test:middleware", 2, time.Minute)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(limiter, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), PrincipalContextKey, Principal{APIKeyID: "test-key"}))

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("first request: expected status %d, got %d", http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("second request: expected status %d, got %d", http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("third request: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}
