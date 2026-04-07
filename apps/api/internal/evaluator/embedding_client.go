package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EmbeddingClient interface {
	CreateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
}

type embeddingConfig struct {
	provider string
	model    string
	apiKey   string
	baseURL  string
}

func getEmbeddingConfig() embeddingConfig {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	if provider == "" {
		provider = "openai"
	}
	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}
	return embeddingConfig{
		provider: provider,
		model:    model,
		apiKey:   os.Getenv("EMBEDDING_API_KEY"),
		baseURL:  os.Getenv("EMBEDDING_BASE_URL"),
	}
}

type HTTPEmbeddingClient struct {
	model    string
	apiKey   string
	baseURL  string
	provider string
}

func NewHTTPEmbeddingClient(model, apiKey, baseURL string) *HTTPEmbeddingClient {
	return NewHTTPEmbeddingClientWithProvider("openai", model, apiKey, baseURL)
}

func NewHTTPEmbeddingClientWithProvider(provider, model, apiKey, baseURL string) *HTTPEmbeddingClient {
	return &HTTPEmbeddingClient{
		provider: provider,
		model:    model,
		apiKey:   apiKey,
		baseURL:  baseURL,
	}
}

func NewHTTPEmbeddingClientFromEnv() *HTTPEmbeddingClient {
	cfg := getEmbeddingConfig()
	return NewHTTPEmbeddingClientWithProvider(cfg.provider, cfg.model, cfg.apiKey, cfg.baseURL)
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []embeddingData `json:"data"`
}

type embeddingData struct {
	Embedding []float64 `json:"embedding"`
}

type ollamaEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *HTTPEmbeddingClient) CreateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if c.baseURL == "" || c.model == "" {
		return nil, nil
	}

	var reqBody interface{}
	var endpoint string

	switch c.provider {
	case "ollama":
		endpoint = c.baseURL + "/api/embeddings"
		if len(texts) == 0 {
			return nil, nil
		}
		reqBody = ollamaEmbeddingRequest{
			Model: c.model,
			Input: texts[0],
		}
	default:
		endpoint = c.baseURL + "/embeddings"
		reqBody = openAIEmbeddingRequest{
			Model: c.model,
			Input: texts,
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	switch c.provider {
	case "ollama":
		var embResp ollamaEmbeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
			return nil, err
		}
		return [][]float64{embResp.Embedding}, nil
	default:
		var embResp openAIEmbeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
			return nil, err
		}
		result := make([][]float64, len(embResp.Data))
		for i, d := range embResp.Data {
			result[i] = d.Embedding
		}
		return result, nil
	}
}
