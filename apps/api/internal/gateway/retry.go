package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
	}
}

func GenerateWithRetry(ctx context.Context, client llm.Client, messages []llm.Message, opts []llm.Option, config RetryConfig) (*llm.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := config.InitialBackoff * time.Duration(1<<(attempt-1))
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := client.Generate(ctx, messages, opts...)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if isNonRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isNonRetryableError(err error) bool {
	errStr := err.Error()

	nonRetryableErrors := []string{
		"invalid API key",
		"authentication failed",
		"invalid model",
		"invalid request",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if contains(errStr, nonRetryable) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func GenerateWithRouterRetry(ctx context.Context, router LLMRouter, messages []llm.Message, opts []llm.Option, config RetryConfig) (*llm.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := config.InitialBackoff * time.Duration(1<<(attempt-1))
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		client, err := router.Route(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Generate(ctx, messages, opts...)
		if err == nil {
			router.RecordSuccess(client.Provider())
			return resp, nil
		}

		lastErr = err
		router.RecordFailure(client.Provider())

		if isNonRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
