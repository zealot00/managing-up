package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type alibabaStreamReader struct {
	body     io.ReadCloser
	scanner  *bufio.Scanner
	model    Model
	provider Provider
	usage    Usage
}

func (r *alibabaStreamReader) Recv() (*StreamChunk, error) {
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var chunk struct {
			Output struct {
				Text string `json:"text"`
			} `json:"output"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
				TotalTokens  int `json:"total_tokens"`
			} `json:"usage"`
			Model        string `json:"model"`
			FinishReason string `json:"finish_reason,omitempty"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Usage.InputTokens > 0 || chunk.Usage.OutputTokens > 0 {
			r.usage.InputTokens = chunk.Usage.InputTokens
			r.usage.OutputTokens = chunk.Usage.OutputTokens
			r.usage.TotalTokens = chunk.Usage.TotalTokens
		}

		if chunk.Output.Text != "" {
			return &StreamChunk{
				Content: chunk.Output.Text,
				Done:    false,
			}, nil
		}

		if chunk.FinishReason != "" {
			return &StreamChunk{
				Content:      "",
				Done:         true,
				FinishReason: chunk.FinishReason,
				Usage:        &r.usage,
				Model:        r.model,
			}, nil
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("stream ended unexpectedly")
}

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
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Combine messages into a prompt
	prompt := ""
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}

	reqBody := map[string]any{
		"model": c.model,
		"input": map[string]string{
			"prompt": prompt,
		},
		"parameters": map[string]any{},
	}
	if options.Temperature > 0 {
		reqBody["parameters"].(map[string]any)["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["parameters"].(map[string]any)["max_tokens"] = options.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/services/aigc/text-generation/generation", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Alibaba API: %w", err)
	}
	defer resp.Body.Close()

	var alibabaResp struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&alibabaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Response{
		Content:      alibabaResp.Output.Text,
		Model:        Model(alibabaResp.Model),
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: alibabaResp.Usage.InputTokens, OutputTokens: alibabaResp.Usage.OutputTokens, TotalTokens: alibabaResp.Usage.TotalTokens},
		FinishReason: "stop",
	}, nil
}

func (c *AlibabaClient) Provider() Provider {
	return ProviderAlibaba
}

func (c *AlibabaClient) Model() Model {
	return c.model
}

func (c *AlibabaClient) GenerateStream(ctx context.Context, messages []Message, opts ...Option) (StreamReader, error) {
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	prompt := ""
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}

	reqBody := map[string]any{
		"model": c.model,
		"input": map[string]string{
			"prompt": prompt,
		},
		"parameters": map[string]any{
			"stream": true,
		},
	}
	if options.Temperature > 0 {
		reqBody["parameters"].(map[string]any)["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["parameters"].(map[string]any)["max_tokens"] = options.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/services/aigc/text-generation/generation", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Alibaba API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Alibaba API returned status %d", resp.StatusCode)
	}

	return &alibabaStreamReader{
		body:     resp.Body,
		scanner:  newLargeBufferScanner(resp.Body),
		model:    c.model,
		provider: c.Provider(),
	}, nil
}
