package llm

import (
	"context"
	"net/http"
	"time"
)

type DeepSeekClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewDeepSeekClient(model Model, apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://api.deepseek.com/v1",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *DeepSeekClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *DeepSeekClient) Provider() Provider {
	return ProviderDeepSeek
}

func (c *DeepSeekClient) Model() Model {
	return c.model
}
