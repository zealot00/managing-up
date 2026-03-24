package llm

import (
	"context"
	"net/http"
	"time"
)

type OllamaClient struct {
	model      Model
	baseURL    string
	httpClient *http.Client
}

func NewOllamaClient(model Model, baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *OllamaClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *OllamaClient) Provider() Provider {
	return ProviderOllama
}

func (c *OllamaClient) Model() Model {
	return c.model
}
