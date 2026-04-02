package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	Role    string `json:"role"`
	Content string `json:"content"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HandleOpenAIChat processes OpenAI /v1/chat/completions requests
func (s *Server) HandleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}

	// Parse request
	var req openAIChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	// Validate model
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	// Streaming not supported yet
	if req.Stream {
		writeError(w, http.StatusNotImplemented, "stream_not_supported", "Streaming is not yet implemented")
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
			Role:    msg.Role,
			Content: msg.Content,
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

	// Resolve upstream provider key:
	// - OpenRouter-like mode: use server-side provider keys
	// - fallback mode: passthrough user-provided key
	upstreamAPIKey := apiKey
	if s.providerKeyResolver != nil {
		if resolved := s.providerKeyResolver.KeyFor(provider); resolved != "" {
			upstreamAPIKey = resolved
		}
	}

	// Create LLM client
	llmClient, err := llm.NewClient(provider, model, upstreamAPIKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "client_creation_failed", fmt.Sprintf("Failed to create LLM client: %v", err))
		return
	}

	// Call LLM
	resp, err := llmClient.Generate(r.Context(), messages, opts...)
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
					Role:    "assistant",
					Content: resp.Content,
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
	// Simple implementation using timestamp + random suffix
	return fmt.Sprintf("%d-%04x", time.Now().UnixNano()%1000000000, time.Now().UnixNano()%0xFFFF)
}
