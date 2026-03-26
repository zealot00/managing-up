package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("failed to call DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	var dsResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(dsResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content:      dsResp.Choices[0].Message.Content,
		Model:        Model(dsResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: dsResp.Usage.PromptTokens, OutputTokens: dsResp.Usage.CompletionTokens, TotalTokens: dsResp.Usage.TotalTokens},
		FinishReason: dsResp.Choices[0].FinishReason,
	}, nil
}

func (c *DeepSeekClient) Provider() Provider {
	return ProviderDeepSeek
}

func (c *DeepSeekClient) Model() Model {
	return c.model
}
