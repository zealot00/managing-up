package evaluator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPEmbeddingClient_OpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("expected path /v1/embeddings, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}

		resp := openAIEmbeddingResponse{
			Data: []embeddingData{
				{Embedding: []float64{0.1, 0.2, 0.3}},
				{Embedding: []float64{0.4, 0.5, 0.6}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPEmbeddingClient("text-embedding-3-small", "test-key", server.URL+"/v1")

	vectors, err := client.CreateEmbeddings(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if vectors[0][0] != 0.1 || vectors[0][1] != 0.2 || vectors[0][2] != 0.3 {
		t.Errorf("unexpected first vector: %v", vectors[0])
	}
	if vectors[1][0] != 0.4 || vectors[1][1] != 0.5 || vectors[1][2] != 0.6 {
		t.Errorf("unexpected second vector: %v", vectors[1])
	}
}

func TestHTTPEmbeddingClient_Ollama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("expected path /api/embeddings, got %s", r.URL.Path)
		}

		var req ollamaEmbeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "nomic-embed-text" {
			t.Errorf("expected model 'nomic-embed-text', got %s", req.Model)
		}
		if req.Input != "hello world" {
			t.Errorf("expected input 'hello world', got %s", req.Input)
		}

		resp := ollamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPEmbeddingClientWithProvider("ollama", "nomic-embed-text", "", server.URL)

	vectors, err := client.CreateEmbeddings(context.Background(), []string{"hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vectors) != 1 {
		t.Fatalf("expected 1 vector, got %d", len(vectors))
	}
	if vectors[0][0] != 0.1 || vectors[0][1] != 0.2 || vectors[0][2] != 0.3 {
		t.Errorf("unexpected vector: %v", vectors[0])
	}
}

func TestHTTPEmbeddingClient_EmptyBaseURL(t *testing.T) {
	client := NewHTTPEmbeddingClient("text-embedding-3-small", "test-key", "")

	vectors, err := client.CreateEmbeddings(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vectors != nil {
		t.Errorf("expected nil vectors for empty base URL, got %v", vectors)
	}
}

func TestHTTPEmbeddingClient_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPEmbeddingClient("text-embedding-3-small", "test-key", server.URL)

	vectors, err := client.CreateEmbeddings(context.Background(), []string{"hello"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
	if vectors != nil {
		t.Errorf("expected nil vectors for server error, got %v", vectors)
	}
}

func TestHTTPEmbeddingClientFromEnv(t *testing.T) {
	client := NewHTTPEmbeddingClientFromEnv()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetEmbeddingConfig(t *testing.T) {
	cfg := getEmbeddingConfig()
	if cfg.provider == "" {
		t.Error("expected non-empty provider")
	}
	if cfg.model == "" {
		t.Error("expected non-empty model")
	}
}
