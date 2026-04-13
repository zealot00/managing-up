package llm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMinimaxGenerate_ReturnsStatusErrorOnNon200(t *testing.T) {
	client := NewMinimaxClient("MiniMax-Text-01", "bad-key")
	client.httpClient = &http.Client{
		Timeout: 2 * time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"error":"invalid api key"}`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}

	_, err := client.Generate(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "status 401") {
		t.Fatalf("unexpected error: %v", err)
	}
}
