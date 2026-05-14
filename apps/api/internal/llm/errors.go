package llm

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported LLM provider")
	ErrInvalidAPIKey       = errors.New("invalid API key")
	ErrInvalidModel        = errors.New("invalid model")
	ErrEmptyResponse       = errors.New("empty response from LLM")
	ErrContextCancelled    = errors.New("context cancelled")
	ErrRateLimited         = errors.New("rate limited")
	ErrAuthentication      = errors.New("authentication failed")
	ErrServerError         = errors.New("server error")
)

// ProviderError is a structured error returned by LLM provider clients.
// It carries the HTTP status code for reliable error classification.
type ProviderError struct {
	Provider   Provider
	StatusCode int
	Message    string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s API returned status %d: %s", e.Provider, e.StatusCode, e.Message)
}

// IsRetryable returns true for transient errors that may succeed on retry.
func (e *ProviderError) IsRetryable() bool {
	return e.StatusCode == 429 || e.StatusCode == 500 || e.StatusCode == 502 || e.StatusCode == 503
}

// IsFallbackable returns true for errors that should trigger provider fallback.
func (e *ProviderError) IsFallbackable() bool {
	return e.IsRetryable()
}

// ErrorClass represents how an error should be handled by the fallback system.
type ErrorClass int

const (
	// ErrorClassNonRetryable means stop immediately (401, 403, 400, etc.)
	ErrorClassNonRetryable ErrorClass = iota
	// ErrorClassRetryable means retry the same provider
	ErrorClassRetryable
	// ErrorClassFallbackable means switch to the next provider in the fallback chain
	ErrorClassFallbackable
)

// ClassifyError determines how an error should be handled for retry/fallback.
// It supports both *ProviderError (typed) and legacy fmt.Errorf string errors.
func ClassifyError(err error) ErrorClass {
	if err == nil {
		return ErrorClassNonRetryable
	}

	// Typed path
	var pe *ProviderError
	if errors.As(err, &pe) {
		if pe.IsFallbackable() {
			return ErrorClassFallbackable
		}
		if pe.IsRetryable() {
			return ErrorClassRetryable
		}
		return ErrorClassNonRetryable
	}

	// Legacy string-based fallback (backward compatibility)
	errStr := strings.ToLower(err.Error())

	// Fallbackable: 429, 500, 502, 503
	fallbackable := []string{"status 429", "status 500", "status 502", "status 503", "rate limit", "overloaded"}
	for _, s := range fallbackable {
		if strings.Contains(errStr, s) {
			return ErrorClassFallbackable
		}
	}

	// Non-retryable: 400, 401, 403, 404, 422
	nonRetryable := []string{
		"invalid api key", "authentication failed", "invalid model",
		"invalid request", "unauthorized", "forbidden",
		"status 400", "status 401", "status 403", "status 404", "status 422",
	}
	for _, s := range nonRetryable {
		if strings.Contains(errStr, s) {
			return ErrorClassNonRetryable
		}
	}

	// Default: retryable (network errors, timeouts, etc.)
	return ErrorClassRetryable
}
