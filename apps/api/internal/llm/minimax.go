package llm

import (
	"context"
	"net/http"
	"time"
)

type MinimaxClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewMinimaxClient(model Model, apiKey string) *MinimaxClient {
	return &MinimaxClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://api.minimax.chat/v1",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *MinimaxClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *MinimaxClient) Provider() Provider {
	return ProviderMinimax
}

func (c *MinimaxClient) Model() Model {
	return c.model
}
