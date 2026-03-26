package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Convert messages to Anthropic format
	var anthropicMessages []map[string]string
	for _, m := range messages {
		anthropicMessages = append(anthropicMessages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	reqBody := map[string]any{
		"model":    c.model,
		"messages": anthropicMessages,
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

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	var anthropicResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
		Model      string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := ""
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &Response{
		Content:      content,
		Model:        Model(anthropicResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: anthropicResp.Usage.InputTokens, OutputTokens: anthropicResp.Usage.OutputTokens, TotalTokens: anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens},
		FinishReason: anthropicResp.StopReason,
	}, nil
}

func (c *AnthropicClient) Provider() Provider {
	return ProviderAnthropic
}

func (c *AnthropicClient) Model() Model {
	return c.model
}
