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

// FallbackConfig controls how GenerateWithFallback behaves per-provider.
type FallbackConfig struct {
	MaxRetriesPerProvider int           // Retries within the same provider before falling back (default: 1)
	Backoff               time.Duration // Backoff between retries within the same provider
}

func DefaultFallbackConfig() FallbackConfig {
	return FallbackConfig{
		MaxRetriesPerProvider: 1,
		Backoff:               500 * time.Millisecond,
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

// GenerateWithFallback tries each client in the fallback chain in order.
// On fallbackable errors (429/503), it moves to the next provider.
// On non-retryable errors (401/403), it stops immediately.
// On success, it returns the response.
func GenerateWithFallback(ctx context.Context, clients []llm.Client, messages []llm.Message, opts []llm.Option, fcfg FallbackConfig, recordSuccess, recordFailure func(llm.Provider)) (*llm.Response, error) {
	if len(clients) == 0 {
		return nil, fmt.Errorf("no clients available for fallback")
	}

	var lastErr error

	for i, client := range clients {
		provider := client.Provider()
		model := client.Model()

		// Quick retry within the same provider
		for attempt := 0; attempt <= fcfg.MaxRetriesPerProvider; attempt++ {
			if attempt > 0 {
				slog.Warn("gateway: intra-provider retry",
					"provider", provider,
					"model", model,
					"attempt", attempt,
					"backoff", fcfg.Backoff)

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(fcfg.Backoff):
				}
			}

			resp, err := client.Generate(ctx, messages, opts...)
			if err == nil {
				if recordSuccess != nil {
					recordSuccess(provider)
				}
				if i > 0 {
					slog.Info("gateway: fallback succeeded",
						"original_provider", clients[0].Provider(),
						"fallback_provider", provider,
						"fallback_model", model)
				}
				return resp, nil
			}

			lastErr = err

			class := llm.ClassifyError(err)
			slog.Warn("gateway: LLM Generate failed",
				"provider", provider,
				"model", model,
				"attempt", attempt,
				"error_class", class,
				"error", err)

			if recordFailure != nil {
				recordFailure(provider)
			}

			switch class {
			case llm.ErrorClassNonRetryable:
				return nil, err
			case llm.ErrorClassFallbackable:
				// Don't waste more retries on this provider — fall back
				goto nextProvider
			case llm.ErrorClassRetryable:
				// Retry within this provider
				continue
			}
		}

	nextProvider:
		slog.Warn("gateway: falling back to next provider",
			"failed_provider", provider,
			"failed_model", model,
			"next_index", i+1,
			"total_providers", len(clients))
	}

	slog.Error("gateway: all providers in fallback chain failed",
		"providers_tried", len(clients),
		"last_error", lastErr)
	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// StreamWithFallback tries each client in the fallback chain for stream initiation.
// Fallback only happens BEFORE any data is sent to the client (at GenerateStream call time).
// Once a stream starts successfully, fallback is no longer possible.
func StreamWithFallback(ctx context.Context, clients []llm.Client, messages []llm.Message, opts []llm.Option, fcfg FallbackConfig, recordSuccess, recordFailure func(llm.Provider)) (llm.StreamReader, error) {
	if len(clients) == 0 {
		return nil, fmt.Errorf("no clients available for fallback")
	}

	var lastErr error

	for i, client := range clients {
		provider := client.Provider()
		model := client.Model()

		// For streaming, only try once per provider (GenerateStream failure = immediate fallback)
		streamReader, err := client.GenerateStream(ctx, messages, opts...)
		if err == nil {
			if recordSuccess != nil {
				recordSuccess(provider)
			}
			if i > 0 {
				slog.Info("gateway: stream fallback succeeded",
					"original_provider", clients[0].Provider(),
					"fallback_provider", provider,
					"fallback_model", model)
			}
			return streamReader, nil
		}

		lastErr = err
		class := llm.ClassifyError(err)

		slog.Warn("gateway: LLM GenerateStream failed",
			"provider", provider,
			"model", model,
			"error_class", class,
			"error", err)

		if recordFailure != nil {
			recordFailure(provider)
		}

		if class == llm.ErrorClassNonRetryable {
			return nil, err
		}

		// Both Fallbackable and Retryable → try next provider for streaming
		slog.Warn("gateway: stream falling back to next provider",
			"failed_provider", provider,
			"next_index", i+1,
			"total_providers", len(clients))
	}

	slog.Error("gateway: all providers in stream fallback chain failed",
		"providers_tried", len(clients),
		"last_error", lastErr)
	return nil, fmt.Errorf("all providers failed: %w", lastErr)
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
