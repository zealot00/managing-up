package llm

import (
	"context"
	"net/http"
	"time"
)

type AzureClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewAzureClient(model Model, apiKey string) *AzureClient {
	return &AzureClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://YOUR_RESOURCE.openai.azure.com/v1",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AzureClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *AzureClient) Provider() Provider {
	return ProviderAzure
}

func (c *AzureClient) Model() Model {
	return c.model
}
