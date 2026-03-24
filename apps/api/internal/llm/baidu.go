package llm

import (
	"context"
	"net/http"
	"time"
)

type BaiduClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewBaiduClient(model Model, apiKey string) *BaiduClient {
	return &BaiduClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://qianfan.baidubce.com/v2",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *BaiduClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return nil, nil
}

func (c *BaiduClient) Provider() Provider {
	return ProviderBaidu
}

func (c *BaiduClient) Model() Model {
	return c.model
}
