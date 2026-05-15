package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// --- Mocks ---

type mockFallbackClient struct {
	provider    llm.Provider
	model       llm.Model
	genResult   *llm.Response
	genErr      error
	streamErr   error
	streamReader llm.StreamReader
}

func (m *mockFallbackClient) Generate(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Response, error) {
	return m.genResult, m.genErr
}

func (m *mockFallbackClient) GenerateStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (llm.StreamReader, error) {
	return m.streamReader, m.streamErr
}

func (m *mockFallbackClient) Provider() llm.Provider { return m.provider }
func (m *mockFallbackClient) Model() llm.Model        { return m.model }

type mockStreamReader struct {
	chunks []*llm.StreamChunk
	index  int
}

func (m *mockStreamReader) Recv() (*llm.StreamChunk, error) {
	if m.index >= len(m.chunks) {
		return nil, errors.New("stream ended")
	}
	chunk := m.chunks[m.index]
	m.index++
	return chunk, nil
}

// --- Tests ---

func TestGenerateWithFallback_SuccessFirstProvider(t *testing.T) {
	t.Parallel()

	client := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genResult: &llm.Response{Content: "hello"},
	}

	clients := []llm.Client{client}
	resp, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hello" {
		t.Errorf("expected hello, got %s", resp.Content)
	}
}

func TestGenerateWithFallback_FallbackOn429(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 429,
			Message:    "rate limited",
		},
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		genResult: &llm.Response{Content: "fallback response"},
	}

	var successProvider, failureProvider llm.Provider
	clients := []llm.Client{client1, client2}
	resp, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(),
		func(p llm.Provider) { successProvider = p },
		func(p llm.Provider) { failureProvider = p },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "fallback response" {
		t.Errorf("expected fallback response, got %s", resp.Content)
	}
	if failureProvider != llm.ProviderOpenAI {
		t.Errorf("expected OpenAI failure recorded, got %s", failureProvider)
	}
	if successProvider != llm.ProviderAnthropic {
		t.Errorf("expected Anthropic success recorded, got %s", successProvider)
	}
}

func TestGenerateWithFallback_FallbackOn503(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 503,
			Message:    "service unavailable",
		},
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		genResult: &llm.Response{Content: "fallback"},
	}

	clients := []llm.Client{client1, client2}
	resp, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("expected fallback, got %s", resp.Content)
	}
}

func TestGenerateWithFallback_StopOn401(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 401,
			Message:    "unauthorized",
		},
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		genResult: &llm.Response{Content: "should not reach here"},
	}

	clients := []llm.Client{client1, client2}
	_, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestGenerateWithFallback_AllProvidersFail(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 429,
			Message:    "rate limited",
		},
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		genErr: &llm.ProviderError{
			Provider:   llm.ProviderAnthropic,
			StatusCode: 503,
			Message:    "service unavailable",
		},
	}

	clients := []llm.Client{client1, client2}
	_, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestGenerateWithFallback_EmptyClientList(t *testing.T) {
	t.Parallel()

	_, err := GenerateWithFallback(context.Background(), nil, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error for empty client list")
	}
}

func TestGenerateWithFallback_LegacyStringError(t *testing.T) {
	t.Parallel()

	// Test fallback with a legacy string error containing "status 429"
	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr:   errors.New("OpenAI API returned status 429"),
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		genResult: &llm.Response{Content: "fallback"},
	}

	clients := []llm.Client{client1, client2}
	resp, err := GenerateWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("expected fallback, got %s", resp.Content)
	}
}

func TestGenerateWithFallback_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		genErr:   errors.New("some error"),
	}

	clients := []llm.Client{client}
	cfg := FallbackConfig{MaxRetriesPerProvider: 1, Backoff: 0}
	_, err := GenerateWithFallback(ctx, clients, nil, nil, cfg, nil, nil)
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

// --- StreamWithFallback Tests ---

func TestStreamWithFallback_SuccessFirstProvider(t *testing.T) {
	t.Parallel()

	sr := &mockStreamReader{chunks: []*llm.StreamChunk{
		{Content: "hello", Done: true},
	}}
	client := &mockFallbackClient{
		provider:     llm.ProviderOpenAI,
		model:        "gpt-4o",
		streamReader: sr,
	}

	clients := []llm.Client{client}
	reader, err := StreamWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader == nil {
		t.Fatal("expected non-nil stream reader")
	}
}

func TestStreamWithFallback_FallbackOn429(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		streamErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 429,
			Message:    "rate limited",
		},
	}
	sr := &mockStreamReader{chunks: []*llm.StreamChunk{
		{Content: "fallback stream", Done: true},
	}}
	client2 := &mockFallbackClient{
		provider:     llm.ProviderAnthropic,
		model:        "claude-sonnet-4",
		streamReader: sr,
	}

	clients := []llm.Client{client1, client2}
	reader, err := StreamWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader == nil {
		t.Fatal("expected non-nil stream reader from fallback")
	}
}

func TestStreamWithFallback_StopOn401(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		streamErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 401,
			Message:    "unauthorized",
		},
	}
	sr := &mockStreamReader{chunks: []*llm.StreamChunk{
		{Content: "should not reach", Done: true},
	}}
	client2 := &mockFallbackClient{
		provider:     llm.ProviderAnthropic,
		model:        "claude-sonnet-4",
		streamReader: sr,
	}

	clients := []llm.Client{client1, client2}
	_, err := StreamWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestStreamWithFallback_AllProvidersFail(t *testing.T) {
	t.Parallel()

	client1 := &mockFallbackClient{
		provider: llm.ProviderOpenAI,
		model:    "gpt-4o",
		streamErr: &llm.ProviderError{
			Provider:   llm.ProviderOpenAI,
			StatusCode: 503,
			Message:    "service unavailable",
		},
	}
	client2 := &mockFallbackClient{
		provider: llm.ProviderAnthropic,
		model:    "claude-sonnet-4",
		streamErr: &llm.ProviderError{
			Provider:   llm.ProviderAnthropic,
			StatusCode: 429,
			Message:    "rate limited",
		},
	}

	clients := []llm.Client{client1, client2}
	_, err := StreamWithFallback(context.Background(), clients, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestStreamWithFallback_EmptyClientList(t *testing.T) {
	t.Parallel()

	_, err := StreamWithFallback(context.Background(), nil, nil, nil, DefaultFallbackConfig(), nil, nil)
	if err == nil {
		t.Fatal("expected error for empty client list")
	}
}

func TestDefaultFallbackConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultFallbackConfig()
	if cfg.MaxRetriesPerProvider != 1 {
		t.Errorf("expected MaxRetriesPerProvider=1, got %d", cfg.MaxRetriesPerProvider)
	}
	if cfg.Backoff == 0 {
		t.Error("expected non-zero backoff")
	}
}
