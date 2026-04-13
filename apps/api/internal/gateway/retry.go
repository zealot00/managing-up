package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

			slog.Warn("gateway: retrying LLM request",
				"attempt", attempt,
				"provider", client.Provider(),
				"model", client.Model(),
				"backoff", backoff,
				"last_error", lastErr)

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
		slog.Warn("gateway: LLM Generate failed",
			"attempt", attempt,
			"provider", client.Provider(),
			"model", client.Model(),
			"error", err)

		if isNonRetryableError(err) {
			slog.Info("gateway: non-retryable error, returning",
				"provider", client.Provider(),
				"model", client.Model(),
				"error", err)
			return nil, err
		}
	}

	slog.Error("gateway: max retries exceeded",
		"provider", client.Provider(),
		"model", client.Model(),
		"last_error", lastErr)
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isNonRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	nonRetryableErrors := []string{
		"invalid api key",
		"authentication failed",
		"invalid model",
		"invalid request",
		"unauthorized",
		"forbidden",
		"status 400",
		"status 401",
		"status 403",
		"status 404",
		"status 422",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(errStr, nonRetryable) {
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

			slog.Warn("gateway: retrying LLM request via router",
				"attempt", attempt,
				"backoff", backoff,
				"last_error", lastErr)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		client, err := router.Route(ctx)
		if err != nil {
			lastErr = err
			slog.Warn("gateway: router.Route failed",
				"attempt", attempt,
				"error", err)
			continue
		}

		resp, err := client.Generate(ctx, messages, opts...)
		if err == nil {
			router.RecordSuccess(client.Provider())
			return resp, nil
		}

		lastErr = err
		router.RecordFailure(client.Provider())
		slog.Warn("gateway: LLM Generate via router failed",
			"attempt", attempt,
			"provider", client.Provider(),
			"model", client.Model(),
			"error", err)

		if isNonRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
