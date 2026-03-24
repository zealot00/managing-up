package llm

import (
	"context"
	"net/http"
	"time"
)

type AlibabaClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewAlibabaClient(model Model, apiKey string) *AlibabaClient {
	return &AlibabaClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://dashscope.aliyuncs.com/api/v1",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *AlibabaClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *AlibabaClient) Provider() Provider {
	return ProviderAlibaba
}

func (c *AlibabaClient) Model() Model {
	return c.model
}
