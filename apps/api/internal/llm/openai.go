package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OpenAIClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIClient(model Model, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://api.openai.com/v1",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
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
	if options.TopP > 0 {
		reqBody["top_p"] = options.TopP
	}
	if len(options.StopWords) > 0 {
		reqBody["stop"] = options.StopWords
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
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	var openAIResp struct {
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

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &Response{
		Content:      openAIResp.Choices[0].Message.Content,
		Model:        Model(openAIResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: openAIResp.Usage.PromptTokens, OutputTokens: openAIResp.Usage.CompletionTokens, TotalTokens: openAIResp.Usage.TotalTokens},
		FinishReason: openAIResp.Choices[0].FinishReason,
	}, nil
}

func (c *OpenAIClient) Provider() Provider {
	return ProviderOpenAI
}

func (c *OpenAIClient) Model() Model {
	return c.model
}
