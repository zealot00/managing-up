package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
)

// MCPResourcesHandler handles MCP resource-related API endpoints.
type MCPResourcesHandler struct {
	registry *executors.MCPRegistry
	repo     MCPServersRepository
}

func NewMCPResourcesHandler(registry *executors.MCPRegistry, repo MCPServersRepository) *MCPResourcesHandler {
	return &MCPResourcesHandler{registry: registry, repo: repo}
}

// ListResources returns the resources available on a specific MCP server.
func (h *MCPResourcesHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverName, ok := h.getServerName(w, r, "/resources")
	if !ok {
		return
	}

	resources, err := h.registry.ListResourcesByServer(r.Context(), serverName)
	if err != nil {
		slog.Warn("failed to list resources", "server", serverName, "error", err)
		writeEnvelope(w, http.StatusOK, "mcp_server_resources", map[string]any{
			"server_name": serverName,
			"resources":   []any{},
		})
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_server_resources", map[string]any{
		"server_name": serverName,
		"resources":   resources,
	})
}

// ReadResource reads a specific resource from an MCP server.
func (h *MCPResourcesHandler) ReadResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverName, ok := h.getServerName(w, r, "/resources/read")
	if !ok {
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.URI == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "uri is required")
		return
	}

	result, err := h.registry.ReadResource(r.Context(), serverName, req.URI)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to read resource: %v", err))
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_resource_read", result)
}

// ListResourceTemplates returns resource templates for a specific server.
func (h *MCPResourcesHandler) ListResourceTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverName, ok := h.getServerName(w, r, "/resources/templates")
	if !ok {
		return
	}

	templates, err := h.registry.ListResourceTemplatesByServer(r.Context(), serverName)
	if err != nil {
		slog.Warn("failed to list resource templates", "server", serverName, "error", err)
		writeEnvelope(w, http.StatusOK, "mcp_server_resource_templates", map[string]any{
			"server_name": serverName,
			"templates":   []any{},
		})
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_server_resource_templates", map[string]any{
		"server_name": serverName,
		"templates":   templates,
	})
}

// Subscribe subscribes to resource changes.
func (h *MCPResourcesHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverName, ok := h.getServerName(w, r, "/resources/subscribe")
	if !ok {
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.registry.SubscribeResource(r.Context(), serverName, req.URI); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to subscribe: %v", err))
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_resource_subscribe", map[string]any{"subscribed": true})
}

// Unsubscribe unsubscribes from resource changes.
func (h *MCPResourcesHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverName, ok := h.getServerName(w, r, "/resources/subscribe")
	if !ok {
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.registry.UnsubscribeResource(r.Context(), serverName, req.URI); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to unsubscribe: %v", err))
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_resource_unsubscribe", map[string]any{"subscribed": false})
}

func (h *MCPResourcesHandler) getServerName(w http.ResponseWriter, r *http.Request, suffix string) (string, bool) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id := strings.TrimSuffix(path, suffix)
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return "", false
	}

	server, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return "", false
	}

	return server.Name, true
}
