package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaClient struct {
	model      Model
	baseURL    string
	httpClient *http.Client
}

func NewOllamaClient(model Model, baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *OllamaClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	reqBody := map[string]any{
		"model":    c.model,
		"messages": messages,
		"stream":   false,
	}
	if options.Temperature > 0 {
		reqBody["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["options"] = map[string]any{
			"num_predict": options.MaxTokens,
		}
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Message struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
		Model       string `json:"model"`
		DoneReason  string `json:"done_reason,omitempty"`
		PromptCount int    `json:"prompt_eval_count,omitempty"`
		EvalCount   int    `json:"eval_count,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Response{
		Content:      ollamaResp.Message.Content,
		Model:        Model(ollamaResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: ollamaResp.PromptCount, OutputTokens: ollamaResp.EvalCount, TotalTokens: ollamaResp.PromptCount + ollamaResp.EvalCount},
		FinishReason: ollamaResp.DoneReason,
	}, nil
}

func (c *OllamaClient) Provider() Provider {
	return ProviderOllama
}

func (c *OllamaClient) Model() Model {
	return c.model
}
