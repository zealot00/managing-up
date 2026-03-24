package llm

import (
	"context"
	"net/http"
	"time"
)

type ZhipuAIClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewZhipuAIClient(model Model, apiKey string) *ZhipuAIClient {
	return &ZhipuAIClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://open.bigmodel.cn/api/paas/v4",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *ZhipuAIClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *ZhipuAIClient) Provider() Provider {
	return ProviderZhipuAI
}

func (c *ZhipuAIClient) Model() Model {
	return c.model
}
