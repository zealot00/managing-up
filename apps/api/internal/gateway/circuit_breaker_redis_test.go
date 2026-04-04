package gateway

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisCircuitBreaker_StateTransitions(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb", 3, 2, 5*time.Second)
	ctx := context.Background()
	key := "test-service"

	if err := cb.Reset(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	state, err := cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected initial state closed, got %s", state)
	}

	allowed, err := cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !allowed {
		t.Error("expected Allow=true in closed state")
	}

	if err := cb.RecordFailure(ctx, key); err != nil {
		t.Fatalf("RecordFailure failed: %v", err)
	}
	if err := cb.RecordFailure(ctx, key); err != nil {
		t.Fatalf("RecordFailure failed: %v", err)
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected still closed after 2 failures, got %s", state)
	}

	if err := cb.RecordFailure(ctx, key); err != nil {
		t.Fatalf("RecordFailure failed: %v", err)
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerOpen {
		t.Errorf("expected open after 3 failures, got %s", state)
	}

	allowed, err = cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if allowed {
		t.Error("expected Allow=false in open state")
	}
}

func TestRedisCircuitBreaker_HalfOpenToClosed(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb:halfopen", 2, 2, 2*time.Second)
	ctx := context.Background()
	key := "test-halfopen-service"

	if err := cb.Reset(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := cb.RecordFailure(ctx, key); err != nil {
			t.Fatalf("RecordFailure %d failed: %v", i, err)
		}
	}

	state, err := cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerOpen {
		t.Errorf("expected open, got %s", state)
	}

	time.Sleep(2500 * time.Millisecond)

	allowed, err := cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !allowed {
		t.Error("expected Allow=true after timeout (half_open)")
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerHalfOpen {
		t.Errorf("expected half_open after timeout, got %s", state)
	}

	if err := cb.RecordSuccess(ctx, key); err != nil {
		t.Fatalf("RecordSuccess failed: %v", err)
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerHalfOpen {
		t.Errorf("expected still half_open after 1 success, got %s", state)
	}

	if err := cb.RecordSuccess(ctx, key); err != nil {
		t.Fatalf("RecordSuccess failed: %v", err)
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected closed after 2 successes in half_open, got %s", state)
	}
}

func TestRedisCircuitBreaker_HalfOpenFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb:halffail", 2, 2, 2*time.Second)
	ctx := context.Background()
	key := "test-halffail-service"

	if err := cb.Reset(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := cb.RecordFailure(ctx, key); err != nil {
			t.Fatalf("RecordFailure %d failed: %v", i, err)
		}
	}

	time.Sleep(2500 * time.Millisecond)

	allowed, err := cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !allowed {
		t.Error("expected Allow=true after timeout")
	}

	if err := cb.RecordSuccess(ctx, key); err != nil {
		t.Fatalf("RecordSuccess failed: %v", err)
	}

	if err := cb.RecordFailure(ctx, key); err != nil {
		t.Fatalf("RecordFailure failed: %v", err)
	}

	state, err := cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerOpen {
		t.Errorf("expected open after failure in half_open, got %s", state)
	}
}

func TestRedisCircuitBreaker_Reset(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb:reset", 3, 2, 5*time.Second)
	ctx := context.Background()
	key := "test-reset-service"

	for i := 0; i < 3; i++ {
		if err := cb.RecordFailure(ctx, key); err != nil {
			t.Fatalf("RecordFailure %d failed: %v", i, err)
		}
	}

	state, err := cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerOpen {
		t.Errorf("expected open, got %s", state)
	}

	if err := cb.Reset(ctx, key); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	state, err = cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed after reset: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected closed after reset, got %s", state)
	}

	allowed, err := cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !allowed {
		t.Error("expected Allow=true after reset")
	}
}

func TestRedisCircuitBreaker_MultipleKeys(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb:multi", 2, 2, 5*time.Second)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := cb.RecordFailure(ctx, "service-a"); err != nil {
			t.Fatalf("RecordFailure service-a %d failed: %v", i, err)
		}
	}

	stateA, _ := cb.State(ctx, "service-a")
	stateB, _ := cb.State(ctx, "service-b")

	if stateA != CircuitBreakerOpen {
		t.Errorf("service-a should be open, got %s", stateA)
	}
	if stateB != CircuitBreakerClosed {
		t.Errorf("service-b should be closed, got %s", stateB)
	}
}

func TestRedisCircuitBreaker_FullIntegration(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis test")
	}

	client := redis.NewClient(&redis.Options{Addr: redisURL})
	defer client.Close()

	cb := NewRedisCircuitBreaker(client, "test:cb:integration", 5, 2, 2*time.Second)
	ctx := context.Background()
	key := "integration-test"

	if err := cb.Reset(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	for i := 0; i < 4; i++ {
		allowed, err := cb.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow %d failed: %v", i, err)
		}
		if !allowed {
			t.Errorf("Allow %d should return true (closed)", i)
		}
	}

	for i := 0; i < 5; i++ {
		if err := cb.RecordFailure(ctx, key); err != nil {
			t.Fatalf("RecordFailure %d failed: %v", i, err)
		}
	}

	allowed, err := cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if allowed {
		t.Error("should not allow after 5 failures")
	}

	time.Sleep(2500 * time.Millisecond)

	allowed, err = cb.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow after timeout failed: %v", err)
	}
	if !allowed {
		t.Error("should allow after timeout")
	}

	for i := 0; i < 2; i++ {
		if err := cb.RecordSuccess(ctx, key); err != nil {
			t.Fatalf("RecordSuccess %d failed: %v", i, err)
		}
	}

	state, err := cb.State(ctx, key)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if state != CircuitBreakerClosed {
		t.Errorf("expected closed after recovery, got %s", state)
	}
}
