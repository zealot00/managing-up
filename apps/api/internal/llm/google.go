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

type googleStreamReader struct {
	body     io.ReadCloser
	scanner  *bufio.Scanner
	model    Model
	provider Provider
	usage    Usage
}

func (r *googleStreamReader) Recv() (*StreamChunk, error) {
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
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason,omitempty"`
			} `json:"candidates"`
			UsageMetadata struct {
				PromptTokenCount     int `json:"promptTokenCount,omitempty"`
				CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
				TotalTokenCount      int `json:"totalTokenCount,omitempty"`
			} `json:"usageMetadata,omitempty"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.UsageMetadata.PromptTokenCount > 0 || chunk.UsageMetadata.CandidatesTokenCount > 0 {
			r.usage.InputTokens = chunk.UsageMetadata.PromptTokenCount
			r.usage.OutputTokens = chunk.UsageMetadata.CandidatesTokenCount
			r.usage.TotalTokens = chunk.UsageMetadata.TotalTokenCount
		}

		if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
			text := chunk.Candidates[0].Content.Parts[0].Text
			finishReason := chunk.Candidates[0].FinishReason

			if finishReason != "" {
				return &StreamChunk{
					Content:      text,
					Done:         true,
					FinishReason: finishReason,
					Usage:        &r.usage,
					Model:        r.model,
				}, nil
			}

			if text != "" {
				return &StreamChunk{
					Content: text,
					Done:    false,
				}, nil
			}
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("stream ended unexpectedly")
}

type GoogleClient struct {
	model      Model
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewGoogleClient(model Model, apiKey string) *GoogleClient {
	return &GoogleClient{
		model:      model,
		apiKey:     apiKey,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *GoogleClient) Generate(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Combine all messages into a single prompt for Gemini
	prompt := ""
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}
	if options.Temperature > 0 {
		reqBody["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["maxOutputTokens"] = options.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google API: %w", err)
	}
	defer resp.Body.Close()

	var googleResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(googleResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	content := ""
	for _, part := range googleResp.Candidates[0].Content.Parts {
		content += part.Text
	}

	return &Response{
		Content:      content,
		Model:        c.model,
		Provider:     c.Provider(),
		Usage:        Usage{InputTokens: googleResp.UsageMetadata.PromptTokenCount, OutputTokens: googleResp.UsageMetadata.CandidatesTokenCount, TotalTokens: googleResp.UsageMetadata.TotalTokenCount},
		FinishReason: "stop",
	}, nil
}

func (c *GoogleClient) Provider() Provider {
	return ProviderGoogle
}

func (c *GoogleClient) Model() Model {
	return c.model
}

func (c *GoogleClient) GenerateStream(ctx context.Context, messages []Message, opts ...Option) (StreamReader, error) {
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	prompt := ""
	for _, m := range messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"stream": true,
	}
	if options.Temperature > 0 {
		reqBody["temperature"] = options.Temperature
	}
	if options.MaxTokens > 0 {
		reqBody["maxOutputTokens"] = options.MaxTokens
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Google API returned status %d", resp.StatusCode)
	}

	return &googleStreamReader{
		body:     resp.Body,
		scanner:  newLargeBufferScanner(resp.Body),
		model:    c.model,
		provider: c.Provider(),
	}, nil
}
