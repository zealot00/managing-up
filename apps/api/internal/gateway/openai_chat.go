package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// OpenAI chat completions request
type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature *float32        `json:"temperature"`
	MaxTokens   *int            `json:"max_tokens"`
	Stream      bool            `json:"stream"`
	Tools       []llm.Tool      `json:"tools,omitempty"` // 新增：工具列表
}

type openAIMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	Name       string         `json:"name,omitempty"`         // 【新增】
	ToolCallID string         `json:"tool_call_id,omitempty"` // 【新增】：关键，接收客户端传来的 ID
	ToolCalls  []llm.ToolCall `json:"tool_calls,omitempty"`
}

// OpenAI chat completions response
type openAIChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Index        int     `json:"index"`
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type message struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []llm.ToolCall `json:"tool_calls,omitempty"` // 新增：工具调用
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HandleOpenAIChat processes OpenAI /v1/chat/completions requests
func (s *Server) HandleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	slog.Info("HandleOpenAIChat: request received",
		"method", r.Method,
		"content_length", r.ContentLength,
		"header_content_type", r.Header.Get("Content-Type"))

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	slog.Info("HandleOpenAIChat: body read", "body_len", len(body), "err", err)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}

	// Parse request
	var req openAIChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("HandleOpenAIChat: JSON parse failed", "err", err, "body_prefix", func() string {
			if len(body) > 200 {
				return string(body[:200])
			}
			return string(body)
		}())
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	slog.Info("HandleOpenAIChat: parsed request", "model", req.Model, "stream", req.Stream, "messages_count", len(req.Messages))

	// Validate model
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	// Extract API key from Authorization header
	apiKey := GetAPIKeyFromContext(r.Context())
	if apiKey == "" {
		apiKey = extractAPIKey(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "missing_api_key", "Authorization header with Bearer token is required")
		return
	}

	// Parse model string to detect provider
	provider, model, err := llm.ParseModelString(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_model", fmt.Sprintf("Failed to parse model: %v", err))
		return
	}

	// Convert messages to llm.Message format
	messages := make([]llm.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,       // 【新增】
			ToolCallID: msg.ToolCallID, // 【新增】：透传 ID 给大模型
			ToolCalls:  msg.ToolCalls,  // 【新增】：透传多轮历史
		}
	}

	// Build LLM options
	var opts []llm.Option
	if req.Temperature != nil {
		opts = append(opts, llm.WithTemperature(*req.Temperature))
	}
	if req.MaxTokens != nil {
		opts = append(opts, llm.WithMaxTokens(*req.MaxTokens))
	}

	if len(req.Tools) > 0 {
		opts = append(opts, llm.WithTools(req.Tools))
	}

	if req.Stream {
		slog.Info("HandleOpenAIChat: calling handleOpenAIChatStream")
		s.handleOpenAIChatStream(w, r, apiKey, provider, model, messages, opts)
		return
	}

	var resp *llm.Response
	err = error(nil)

	if s.router != nil {
		upstreamAPIKey := apiKey
		if s.providerKeyResolver != nil {
			principal := GetPrincipalFromContext(r.Context())
			userID := ""
			if principal != nil {
				userID = principal.UserID
			}
			if resolved := s.providerKeyResolver.KeyFor(userID, provider); resolved != "" {
				upstreamAPIKey = resolved
			}
		}
		llmClient, createErr := llm.NewClient(provider, model, upstreamAPIKey)
		if createErr != nil {
			writeError(w, http.StatusInternalServerError, "client_creation_failed", fmt.Sprintf("Failed to create LLM client: %v", createErr))
			return
		}
		resp, err = GenerateWithRetry(r.Context(), llmClient, messages, opts, DefaultRetryConfig())
	} else {
		upstreamAPIKey := apiKey
		if s.providerKeyResolver != nil {
			principal := GetPrincipalFromContext(r.Context())
			userID := ""
			if principal != nil {
				userID = principal.UserID
			}
			if resolved := s.providerKeyResolver.KeyFor(userID, provider); resolved != "" {
				upstreamAPIKey = resolved
			}
		}

		llmClient, createErr := llm.NewClient(provider, model, upstreamAPIKey)
		if createErr != nil {
			writeError(w, http.StatusInternalServerError, "client_creation_failed", fmt.Sprintf("Failed to create LLM client: %v", createErr))
			return
		}

		resp, err = GenerateWithRetry(r.Context(), llmClient, messages, opts, DefaultRetryConfig())
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "generation_failed", fmt.Sprintf("LLM generation failed: %v", err))
		return
	}

	// Build OpenAI response
	chatResp := openAIChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", generateID()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   string(resp.Model),
		Choices: []choice{
			{
				Index: 0,
				Message: message{
					Role:      "assistant",
					Content:   resp.Content,
					ToolCalls: resp.ToolCalls,
				},
				FinishReason: resp.FinishReason,
			},
		},
		Usage: usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chatResp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_failed", "Failed to encode response")
		return
	}

	if s.usageRecorder != nil {
		principal := GetPrincipalFromContext(r.Context())
		if principal != nil {
			cost := CalculateCost(string(model), resp.Usage.InputTokens, resp.Usage.OutputTokens)
			_ = s.usageRecorder.RecordUsage(r.Context(), UsageRecord{
				APIKeyID:         principal.APIKeyID,
				UserID:           principal.UserID,
				Username:         principal.Username,
				Provider:         provider,
				Model:            model,
				Endpoint:         "/v1/chat/completions",
				PromptTokens:     resp.Usage.InputTokens,
				CompletionTokens: resp.Usage.OutputTokens,
				TotalTokens:      resp.Usage.TotalTokens,
				Cost:             cost,
			})
		}
	}
}

// extractBearerToken extracts the token from "Bearer <token>" format
func extractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// generateID generates a simple random ID
func generateID() string {
	return fmt.Sprintf("%d-%04x", time.Now().UnixNano()%1000000000, time.Now().UnixNano()%0xFFFF)
}

func (s *Server) handleOpenAIChatStream(w http.ResponseWriter, r *http.Request, apiKey string, provider llm.Provider, model llm.Model, messages []llm.Message, opts []llm.Option) {
	slog.Info("handleOpenAIChatStream: started", "provider", provider, "model", model, "messages_count", len(messages))
	upstreamAPIKey := apiKey
	if s.providerKeyResolver != nil {
		principal := GetPrincipalFromContext(r.Context())
		userID := ""
		if principal != nil {
			userID = principal.UserID
		}
		if resolved := s.providerKeyResolver.KeyFor(userID, provider); resolved != "" {
			upstreamAPIKey = resolved
		}
	}

	llmClient, err := llm.NewClient(provider, model, upstreamAPIKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "client_creation_failed", fmt.Sprintf("Failed to create LLM client: %v", err))
		return
	}

	streamReader, err := llmClient.GenerateStream(r.Context(), messages, opts...)
	slog.Info("handleOpenAIChatStream: GenerateStream called", "err", err)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stream_failed", fmt.Sprintf("Failed to start stream: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream_not_supported", "Streaming not supported")
		return
	}

	id := fmt.Sprintf("chatcmpl-%s", generateID())
	var totalPromptTokens, totalCompletionTokens int
	var totalContentLen int

	const (
		flushTimeThreshold = 50 * time.Millisecond
		flushSizeThreshold = 20
	)

	var textBuffer strings.Builder
	lastFlushTime := time.Now()
	isFirstToken := true

	flushBuffer := func(content string, toolCalls []llm.ToolCall, finishReason string) {
		delta := map[string]any{}
		if content != "" {
			delta["content"] = content
		}
		if len(toolCalls) > 0 {
			delta["tool_calls"] = toolCalls
		}

		if len(delta) == 0 && finishReason == "" {
			return
		}

		var fr any = finishReason
		if finishReason == "" {
			fr = nil
		}

		chunk := map[string]any{
			"id":      id,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   string(model),
			"choices": []map[string]any{
				{
					"index":         0,
					"delta":         delta,
					"finish_reason": fr,
				},
			},
		}

		chunkJSON, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		lastFlushTime = time.Now()
	}

	sendDone := func(finishReason string) {
		doneData := map[string]any{
			"id":      id,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   string(model),
			"choices": []map[string]any{
				{
					"index":         0,
					"delta":         map[string]any{},
					"finish_reason": finishReason,
				},
			},
		}
		doneJSON, _ := json.Marshal(doneData)
		fmt.Fprintf(w, "data: %s\n\n", doneJSON)
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}

	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			slog.Warn("handleOpenAIChatStream: streamReader.Recv error",
				"error", err)
			if textBuffer.Len() > 0 {
				flushBuffer(textBuffer.String(), nil, "")
			}
			break
		}

		slog.Debug("handleOpenAIChatStream: received chunk",
			"content_len", len(chunk.Content),
			"done", chunk.Done,
			"finish_reason", chunk.FinishReason,
			"tool_calls_len", len(chunk.ToolCalls),
			"total_content_len_so_far", totalContentLen)

		if chunk.Done {
			totalContentLen += len(chunk.Content)

			if chunk.Usage != nil {
				totalPromptTokens = chunk.Usage.InputTokens
				totalCompletionTokens = chunk.Usage.OutputTokens
			}

			if textBuffer.Len() > 0 || len(chunk.ToolCalls) > 0 {
				flushBuffer(textBuffer.String(), chunk.ToolCalls, "")
			}
			sendDone(chunk.FinishReason)

			slog.Info("handleOpenAIChatStream: stream completed",
				"total_content_len", totalContentLen,
				"prompt_tokens", totalPromptTokens,
				"completion_tokens", totalCompletionTokens)
			break
		}

		totalContentLen += len(chunk.Content)

		hasToolCalls := len(chunk.ToolCalls) > 0
		isDone := chunk.Done || chunk.FinishReason != ""
		timeReached := time.Since(lastFlushTime) > flushTimeThreshold
		sizeReached := textBuffer.Len() > flushSizeThreshold

		if chunk.Content != "" {
			textBuffer.WriteString(chunk.Content)
		}

		shouldFlush := isFirstToken || hasToolCalls || isDone || timeReached || sizeReached

		if shouldFlush && textBuffer.Len() > 0 {
			flushBuffer(textBuffer.String(), chunk.ToolCalls, "")
			textBuffer.Reset()
			isFirstToken = false
		} else if hasToolCalls {
			flushBuffer("", chunk.ToolCalls, "")
		}
	}

	if s.usageRecorder != nil {
		principal := GetPrincipalFromContext(r.Context())
		if principal != nil {
			cost := CalculateCost(string(model), totalPromptTokens, totalCompletionTokens)
			_ = s.usageRecorder.RecordUsage(r.Context(), UsageRecord{
				APIKeyID:         principal.APIKeyID,
				UserID:           principal.UserID,
				Username:         principal.Username,
				Provider:         provider,
				Model:            model,
				Endpoint:         "/v1/chat/completions",
				PromptTokens:     totalPromptTokens,
				CompletionTokens: totalCompletionTokens,
				TotalTokens:      totalPromptTokens + totalCompletionTokens,
				Cost:             cost,
			})
		}
	}
}
