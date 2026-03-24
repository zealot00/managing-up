package llm

import (
	"context"
	"net/http"
	"time"
)

type GoogleClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewGoogleClient(model Model, apiKey string) *GoogleClient {
	return &GoogleClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *GoogleClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *GoogleClient) Provider() Provider {
	return ProviderGoogle
}

func (c *GoogleClient) Model() Model {
	return c.model
}
