package llm

import (
	"errors"
	"testing"
)

func TestProviderError_Error(t *testing.T) {
	t.Parallel()

	pe := &ProviderError{
		Provider:   ProviderOpenAI,
		StatusCode: 429,
		Message:    "rate limit exceeded",
	}

	errStr := pe.Error()
	if errStr != "openai API returned status 429: rate limit exceeded" {
		t.Errorf("unexpected error string: %s", errStr)
	}
}

func TestProviderError_IsRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		statusCode int
		expected   bool
	}{
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{200, false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			pe := &ProviderError{StatusCode: tt.statusCode}
			if got := pe.IsRetryable(); got != tt.expected {
				t.Errorf("IsRetryable() for status %d = %v, want %v", tt.statusCode, got, tt.expected)
			}
		})
	}
}

func TestProviderError_IsFallbackable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		statusCode int
		expected   bool
	}{
		{429, true},
		{503, true},
		{500, true},
		{502, true},
		{400, false},
		{401, false},
		{403, false},
	}

	for _, tt := range tests {
		pe := &ProviderError{StatusCode: tt.statusCode}
		if got := pe.IsFallbackable(); got != tt.expected {
			t.Errorf("IsFallbackable() for status %d = %v, want %v", tt.statusCode, got, tt.expected)
		}
	}
}

func TestClassifyError_TypedProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected ErrorClass
	}{
		{"nil", nil, ErrorClassNonRetryable},
		{"429 rate limit", &ProviderError{StatusCode: 429}, ErrorClassFallbackable},
		{"503 service unavailable", &ProviderError{StatusCode: 503}, ErrorClassFallbackable},
		{"500 internal error", &ProviderError{StatusCode: 500}, ErrorClassFallbackable},
		{"502 bad gateway", &ProviderError{StatusCode: 502}, ErrorClassFallbackable},
		{"401 unauthorized", &ProviderError{StatusCode: 401}, ErrorClassNonRetryable},
		{"403 forbidden", &ProviderError{StatusCode: 403}, ErrorClassNonRetryable},
		{"400 bad request", &ProviderError{StatusCode: 400}, ErrorClassNonRetryable},
		{"404 not found", &ProviderError{StatusCode: 404}, ErrorClassNonRetryable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyError(tt.err)
			if got != tt.expected {
				t.Errorf("ClassifyError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClassifyError_LegacyStringErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected ErrorClass
	}{
		{"status 429 string", errors.New("OpenAI API returned status 429"), ErrorClassFallbackable},
		{"status 503 string", errors.New("Anthropic API returned status 503"), ErrorClassFallbackable},
		{"status 500 string", errors.New("server error status 500"), ErrorClassFallbackable},
		{"rate limit string", errors.New("rate limit exceeded"), ErrorClassFallbackable},
		{"status 401 string", errors.New("API returned status 401: unauthorized"), ErrorClassNonRetryable},
		{"status 403 string", errors.New("API returned status 403: forbidden"), ErrorClassNonRetryable},
		{"status 400 string", errors.New("API returned status 400: bad request"), ErrorClassNonRetryable},
		{"forbidden string", errors.New("access forbidden"), ErrorClassNonRetryable},
		{"unauthorized string", errors.New("unauthorized access"), ErrorClassNonRetryable},
		{"network error", errors.New("connection refused"), ErrorClassRetryable},
		{"timeout error", errors.New("context deadline exceeded"), ErrorClassRetryable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyError(tt.err)
			if got != tt.expected {
				t.Errorf("ClassifyError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestProviderError_ImplementsError(t *testing.T) {
	t.Parallel()

	var err error = &ProviderError{
		Provider:   ProviderAnthropic,
		StatusCode: 429,
		Message:    "rate limited",
	}

	// Verify it implements the error interface
	_ = err.Error()

	// Verify errors.As works
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatal("errors.As should find ProviderError")
	}
	if pe.StatusCode != 429 {
		t.Errorf("expected StatusCode 429, got %d", pe.StatusCode)
	}
	if pe.Provider != ProviderAnthropic {
		t.Errorf("expected Provider anthropic, got %s", pe.Provider)
	}
}
