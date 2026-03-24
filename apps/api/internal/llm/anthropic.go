package llm

import (
	"context"
	"net/http"
	"time"
)

type AnthropicClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewAnthropicClient(model Model, apiKey string) *AnthropicClient {
	return &AnthropicClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://api.anthropic.com/v1",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AnthropicClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *AnthropicClient) Provider() Provider {
	return ProviderAnthropic
}

func (c *AnthropicClient) Model() Model {
	return c.model
}
