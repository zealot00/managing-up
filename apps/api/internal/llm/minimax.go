package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/text/chatcompletion_v2", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MiniMax API: %w", err)
	}
	defer resp.Body.Close()

	var miniResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&miniResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(miniResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content:      miniResp.Choices[0].Message.Content,
		Model:        Model(miniResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: miniResp.Usage.PromptTokens, OutputTokens: miniResp.Usage.CompletionTokens, TotalTokens: miniResp.Usage.TotalTokens},
		FinishReason: miniResp.Choices[0].FinishReason,
	}, nil
}

func (c *MinimaxClient) Provider() Provider {
	return ProviderMinimax
}

func (c *MinimaxClient) Model() Model {
	return c.model
}

func (c *MinimaxClient) GenerateStream(ctx context.Context, messages []Message, opts ...Option) (StreamReader, error) {
	return nil, fmt.Errorf("streaming is not yet supported for MiniMax")
}
