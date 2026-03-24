package llm

import (
	"context"
	"net/http"
	"time"
)

type OpenAIClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIClient(model Model, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://api.openai.com/v1",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *OpenAIClient) Provider() Provider {
	return ProviderOpenAI
}

func (c *OpenAIClient) Model() Model {
	return c.model
}
