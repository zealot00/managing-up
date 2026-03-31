package gateway

import (
	"encoding/json"
	"net/http"
	"time"
)

// OpenAI Models list response format
type modelsResponse struct {
	Object string  `json:"object"`
	Data   []model `json:"data"`
}

type model struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Created  int64  `json:"created"`
	OwnedBy  string `json:"owned_by"`
	Provider string `json:"provider,omitempty"`
}

// HandleModels handles GET /v1/models - list available models
func (s *Server) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// Return list of supported models
	models := []model{
		// OpenAI models
		{ID: "gpt-4o", Object: "model", Created: time.Now().Unix(), OwnedBy: "openai", Provider: "openai"},
		{ID: "gpt-4o-mini", Object: "model", Created: time.Now().Unix(), OwnedBy: "openai", Provider: "openai"},
		{ID: "gpt-4-turbo", Object: "model", Created: time.Now().Unix(), OwnedBy: "openai", Provider: "openai"},
		{ID: "o1-preview", Object: "model", Created: time.Now().Unix(), OwnedBy: "openai", Provider: "openai"},
		{ID: "o1-mini", Object: "model", Created: time.Now().Unix(), OwnedBy: "openai", Provider: "openai"},
		// Anthropic models
		{ID: "claude-sonnet-4-20250514", Object: "model", Created: time.Now().Unix(), OwnedBy: "anthropic", Provider: "anthropic"},
		{ID: "claude-opus-4-20250514", Object: "model", Created: time.Now().Unix(), OwnedBy: "anthropic", Provider: "anthropic"},
		{ID: "claude-haiku-3-20250722", Object: "model", Created: time.Now().Unix(), OwnedBy: "anthropic", Provider: "anthropic"},
		// Google models
		{ID: "gemini-2.0-flash", Object: "model", Created: time.Now().Unix(), OwnedBy: "google", Provider: "google"},
		{ID: "gemini-1.5-flash", Object: "model", Created: time.Now().Unix(), OwnedBy: "google", Provider: "google"},
		// Ollama models (local)
		{ID: "llama3", Object: "model", Created: time.Now().Unix(), OwnedBy: "ollama", Provider: "ollama"},
		{ID: "mistral", Object: "model", Created: time.Now().Unix(), OwnedBy: "ollama", Provider: "ollama"},
		{ID: "qwen2.5", Object: "model", Created: time.Now().Unix(), OwnedBy: "ollama", Provider: "ollama"},
		// Minimax models
		{ID: "abab6.5s-chat", Object: "model", Created: time.Now().Unix(), OwnedBy: "minimax", Provider: "minimax"},
		{ID: "MiniMax-Text-01", Object: "model", Created: time.Now().Unix(), OwnedBy: "minimax", Provider: "minimax"},
		{ID: "MiniMax-Text-01-Mini", Object: "model", Created: time.Now().Unix(), OwnedBy: "minimax", Provider: "minimax"},
		// Zhipu AI models
		{ID: "glm-4", Object: "model", Created: time.Now().Unix(), OwnedBy: "zhipuai", Provider: "zhipuai"},
		{ID: "glm-4-flash", Object: "model", Created: time.Now().Unix(), OwnedBy: "zhipuai", Provider: "zhipuai"},
		{ID: "glm-4-plus", Object: "model", Created: time.Now().Unix(), OwnedBy: "zhipuai", Provider: "zhipuai"},
		{ID: "glm-4v", Object: "model", Created: time.Now().Unix(), OwnedBy: "zhipuai", Provider: "zhipuai"},
		{ID: "glm-3-flash", Object: "model", Created: time.Now().Unix(), OwnedBy: "zhipuai", Provider: "zhipuai"},
		// DeepSeek models
		{ID: "deepseek-chat", Object: "model", Created: time.Now().Unix(), OwnedBy: "deepseek", Provider: "deepseek"},
		{ID: "deepseek-coder", Object: "model", Created: time.Now().Unix(), OwnedBy: "deepseek", Provider: "deepseek"},
		// Baidu ERNIE models
		{ID: "ernie-4.0-8k-latest", Object: "model", Created: time.Now().Unix(), OwnedBy: "baidu", Provider: "baidu"},
		{ID: "ernie-3.5-8k", Object: "model", Created: time.Now().Unix(), OwnedBy: "baidu", Provider: "baidu"},
		{ID: "ernie-3.5-8k-view", Object: "model", Created: time.Now().Unix(), OwnedBy: "baidu", Provider: "baidu"},
		{ID: "ernie-4.0-8k", Object: "model", Created: time.Now().Unix(), OwnedBy: "baidu", Provider: "baidu"},
		// Alibaba Qwen models
		{ID: "qwen-max", Object: "model", Created: time.Now().Unix(), OwnedBy: "alibaba", Provider: "alibaba"},
		{ID: "qwen-plus", Object: "model", Created: time.Now().Unix(), OwnedBy: "alibaba", Provider: "alibaba"},
		{ID: "qwen-turbo", Object: "model", Created: time.Now().Unix(), OwnedBy: "alibaba", Provider: "alibaba"},
		{ID: "qwen-max-long", Object: "model", Created: time.Now().Unix(), OwnedBy: "alibaba", Provider: "alibaba"},
	}

	resp := modelsResponse{
		Object: "list",
		Data:   models,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_failed", "Failed to encode response")
		return
	}
}
