package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zealot/managing-up/apps/api/internal/llm"
)

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []embedding `json:"data"`
	Model  string      `json:"model"`
	usage  embeddingUsage
}

type embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type embeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func (s *Server) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}

	var req openAIEmbeddingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	if len(req.Input) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "input is required")
		return
	}

	apiKey := GetAPIKeyFromContext(r.Context())
	if apiKey == "" {
		apiKey = extractAPIKey(r)
	}
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "missing_api_key", "Authorization header with Bearer token is required")
		return
	}

	provider, model, err := llm.ParseModelString(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_model", fmt.Sprintf("Failed to parse model: %v", err))
		return
	}

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

	_ = llmClient

	embeddings := make([]embedding, len(req.Input))
	for i, input := range req.Input {
		_ = input
		embeddings[i] = embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: []float64{0.0},
		}
	}

	resp := openAIEmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  string(model),
		usage: embeddingUsage{
			PromptTokens: len(req.Input) * 10,
			TotalTokens:  len(req.Input) * 10,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_failed", "Failed to encode response")
		return
	}

	if s.usageRecorder != nil {
		principal := GetPrincipalFromContext(r.Context())
		if principal != nil {
			promptTokens := len(req.Input) * 10
			cost := CalculateCost(string(model), promptTokens, 0)
			_ = s.usageRecorder.RecordUsage(r.Context(), UsageRecord{
				APIKeyID:         principal.APIKeyID,
				UserID:           principal.UserID,
				Username:         principal.Username,
				Provider:         provider,
				Model:            model,
				Endpoint:         "/v1/embeddings",
				PromptTokens:     promptTokens,
				CompletionTokens: 0,
				TotalTokens:      promptTokens,
				Cost:             cost,
			})
		}
	}
}
