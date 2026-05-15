package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
)

// MCPPromptsHandler handles MCP prompt-related API endpoints.
type MCPPromptsHandler struct {
	registry *executors.MCPRegistry
	repo     MCPServersRepository
}

func NewMCPPromptsHandler(registry *executors.MCPRegistry, repo MCPServersRepository) *MCPPromptsHandler {
	return &MCPPromptsHandler{registry: registry, repo: repo}
}

// ListPrompts returns the prompts available on a specific MCP server.
func (h *MCPPromptsHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id := strings.TrimSuffix(path, "/prompts")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return
	}

	server, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	prompts, err := h.registry.ListPromptsByServer(r.Context(), server.Name)
	if err != nil {
		writeEnvelope(w, http.StatusOK, "mcp_server_prompts", map[string]any{
			"server_id":   server.ID,
			"server_name": server.Name,
			"prompts":     []any{},
		})
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_server_prompts", map[string]any{
		"server_id":   server.ID,
		"server_name": server.Name,
		"prompts":     prompts,
	})
}

// GetPrompt retrieves a specific prompt from an MCP server.
func (h *MCPPromptsHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	parts := strings.SplitN(strings.TrimPrefix(path, ""), "/", 3)
	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID and prompt name are required.")
		return
	}
	id := parts[0]
	// parts[1] is "prompts"
	promptName := ""
	if len(parts) >= 3 {
		promptName = parts[2]
	}
	if promptName == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "prompt name is required in URL path.")
		return
	}

	server, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	var req struct {
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Arguments = nil
	}

	result, err := h.registry.GetPrompt(r.Context(), server.Name, promptName, req.Arguments)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to get prompt: %v", err))
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_prompt_get", result)
}
