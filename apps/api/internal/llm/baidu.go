package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	reqBody := map[string]any{
		"model":    c.model,
		"messages": messages,
	}
	if options.Temperature > 0 {
		reqBody["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["max_tokens"] = options.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Baidu API: %w", err)
	}
	defer resp.Body.Close()

	var baiduResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&baiduResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(baiduResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content:      baiduResp.Choices[0].Message.Content,
		Model:        Model(baiduResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: baiduResp.Usage.PromptTokens, OutputTokens: baiduResp.Usage.CompletionTokens, TotalTokens: baiduResp.Usage.TotalTokens},
		FinishReason: baiduResp.Choices[0].FinishReason,
	}, nil
}

func (c *BaiduClient) Provider() Provider {
	return ProviderBaidu
}

func (c *BaiduClient) Model() Model {
	return c.model
}
