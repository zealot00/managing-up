package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

// Anthropic messages request
type anthropicMessageRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float32           `json:"temperature"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Anthropic messages response
type anthropicMessageResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
	Model   string             `json:"model"`
	Usage   anthropicUsage     `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// HandleAnthropicMessages processes Anthropic /v1/messages requests
func (s *Server) HandleAnthropicMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	anthropicVersion := r.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "anthropic-version header is required")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}

	var req anthropicMessageRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	if req.MaxTokens <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "max_tokens must be greater than 0")
		return
	}

	apiKey := r.Header.Get("x-api-key")
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "missing_api_key", "x-api-key header is required")
		return
	}

	provider, model, err := llm.ParseModelString(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_model", fmt.Sprintf("Failed to parse model: %v", err))
		return
	}

	messages := make([]llm.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	var opts []llm.Option
	if req.Temperature != nil {
		opts = append(opts, llm.WithTemperature(*req.Temperature))
	}
	opts = append(opts, llm.WithMaxTokens(req.MaxTokens))

	llmClient, err := llm.NewClient(provider, model, apiKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "client_creation_failed", fmt.Sprintf("Failed to create LLM client: %v", err))
		return
	}

	resp, err := llmClient.Generate(r.Context(), messages, opts...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generation_failed", fmt.Sprintf("LLM generation failed: %v", err))
		return
	}

	anthropicResp := anthropicMessageResponse{
		ID:   fmt.Sprintf("msg_%s", generateID()),
		Type: "message",
		Role: "assistant",
		Content: []anthropicContent{
			{
				Type: "text",
				Text: resp.Content,
			},
		},
		Model: string(resp.Model),
		Usage: anthropicUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(anthropicResp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_failed", "Failed to encode response")
		return
	}
}
