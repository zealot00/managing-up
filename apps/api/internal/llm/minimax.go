package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type minimaxStreamReader struct {
	body     io.ReadCloser
	scanner  *bufio.Scanner
	model    Model
	provider Provider
	usage    Usage
}

func (r *minimaxStreamReader) Recv() (*StreamChunk, error) {
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			slog.Info("minimax stream: received [DONE]")
			return &StreamChunk{
				Done:         true,
				FinishReason: "stop",
				Usage:        &r.usage,
				Model:        r.model,
			}, nil
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Error("minimax stream: failed to unmarshal chunk",
				"error", err,
				"data_len", len(data),
				"data_prefix", func() string {
					if len(data) > 100 {
						return data[:100]
					}
					return data
				}())
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			finishReason := chunk.Choices[0].FinishReason

			slog.Debug("minimax stream: chunk",
				"content_len", len(content),
				"finish_reason", finishReason)

			if chunk.Usage != nil {
				r.usage = Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
					TotalTokens:  chunk.Usage.TotalTokens,
				}
			}

			if finishReason != "" {
				return &StreamChunk{
					Content:      content,
					Done:         true,
					FinishReason: finishReason,
					Usage:        &r.usage,
					Model:        r.model,
				}, nil
			}

			if content != "" {
				return &StreamChunk{
					Content: content,
					Done:    false,
				}, nil
			}
		}
	}

	if err := r.scanner.Err(); err != nil {
		slog.Error("minimax stream: scanner error",
			"error", err)
		return nil, err
	}

	slog.Warn("minimax stream: ended without [DONE]")
	return nil, fmt.Errorf("stream ended without [DONE]")
}

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
	var options GenerateOptions
	for _, opt := range opts {
		opt(&options)
	}

	reqBody := map[string]any{
		"model":    c.model,
		"messages": messages,
		"stream":   true,
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

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("MiniMax API returned status %d", resp.StatusCode)
	}

	return &minimaxStreamReader{
		body:     resp.Body,
		scanner:  newLargeBufferScanner(resp.Body),
		model:    c.model,
		provider: c.Provider(),
	}, nil
}
