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
	slog.Info("minimax stream: Recv() called, starting scan")
	for r.scanner.Scan() {
		line := r.scanner.Text()
		slog.Info("minimax stream: line received", "line_len", len(line), "line_prefix", func() string {
			if len(line) > 100 {
				return line[:100]
			}
			return line
		}())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			// 【修复】：拦截并解析 Minimax 的纯 JSON 错误格式
			var errResp struct {
				BaseResp struct {
					StatusCode int    `json:"status_code"`
					StatusMsg  string `json:"status_msg"`
				} `json:"base_resp"`
			}
			// 尝试解析，如果有 base_resp 且 status_code 非 0，说明 API 报错了
			if err := json.Unmarshal([]byte(line), &errResp); err == nil && errResp.BaseResp.StatusCode != 0 {
				slog.Error("minimax stream: api error received",
					"status_code", errResp.BaseResp.StatusCode,
					"status_msg", errResp.BaseResp.StatusMsg)
				// 向上层抛出明确的错误，而不是静默停止
				return nil, fmt.Errorf("minimax api error: [%d] %s", errResp.BaseResp.StatusCode, errResp.BaseResp.StatusMsg)
			}

			// 如果不是错误 JSON，再 continue 跳过
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		slog.Info("minimax stream: trimmed data", "data", func() string {
			if len(data) > 100 {
				return data[:100]
			}
			return data
		}())
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
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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
					if len(data) > 200 {
						return data[:200]
					}
					return data
				}())
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			toolCalls := chunk.Choices[0].Delta.ToolCalls
			finishReason := chunk.Choices[0].FinishReason

			// 【新增逻辑】：强制兜底，确保返回给前端的 ToolCall 始终带有 index
			for i := range toolCalls {
				if toolCalls[i].Index == nil {
					idx := i // 防止闭包变量地址问题
					toolCalls[i].Index = &idx
				}

				if toolCalls[i].Function != nil && toolCalls[i].Function.Name != "" {
					if toolCalls[i].ID == "" {
						toolCalls[i].ID = fmt.Sprintf("call_mm_%d_%d", time.Now().UnixNano(), i)
					}
					if toolCalls[i].Type == "" {
						toolCalls[i].Type = "function"
					}
				}
			}

			slog.Debug("minimax stream: chunk",
				"content_len", len(content),
				"tool_calls_len", len(toolCalls),
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
					ToolCalls:    toolCalls,
					Done:         true,
					FinishReason: finishReason,
					Usage:        &r.usage,
					Model:        r.model,
				}, nil
			}

			if content != "" || len(toolCalls) > 0 {
				return &StreamChunk{
					Content:   content,
					ToolCalls: toolCalls,
					Done:      false,
				}, nil
			}
		}
	}

	if err := r.scanner.Err(); err != nil {
		slog.Error("minimax stream: scanner error",
			"error", err)
		return nil, err
	}

	slog.Warn("minimax stream: ended without [DONE], treating as stop")
	return &StreamChunk{
		Done:         true,
		FinishReason: "stop",
		Usage:        &r.usage,
		Model:        r.model,
	}, nil
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

	if len(options.Tools) > 0 {
		reqBody["tools"] = options.Tools
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

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MiniMax API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var miniResp struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls,omitempty"`
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
		ToolCalls:    miniResp.Choices[0].Message.ToolCalls,
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
	slog.Info("MinimaxClient: GenerateStream called", "model", c.model, "messages_count", len(messages))
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

	if len(options.Tools) > 0 {
		reqBody["tools"] = options.Tools
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
	slog.Info("MinimaxClient: HTTP response", "status", resp.StatusCode, "err", err)
	if err != nil {
		return nil, fmt.Errorf("failed to call MiniMax API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		slog.Error("MinimaxClient: non-200 response", "status", resp.StatusCode, "body", string(bodyBytes))
		return nil, fmt.Errorf("MiniMax API returned status %d", resp.StatusCode)
	}

	return &minimaxStreamReader{
		body:     resp.Body,
		scanner:  newLargeBufferScanner(resp.Body),
		model:    c.model,
		provider: c.Provider(),
	}, nil
}
