package handlers

import (
	"net/http"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
)

// MCPHealthHandler handles MCP health-related API endpoints.
type MCPHealthHandler struct {
	registry *executors.MCPRegistry
	repo     MCPServersRepository
}

func NewMCPHealthHandler(registry *executors.MCPRegistry, repo MCPServersRepository) *MCPHealthHandler {
	return &MCPHealthHandler{registry: registry, repo: repo}
}

// HealthCheckAll returns the health status of all registered MCP servers.
func (h *MCPHealthHandler) HealthCheckAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if h.registry == nil {
		writeEnvelope(w, http.StatusOK, "mcp_health", map[string]any{
			"servers": []any{},
		})
		return
	}

	results := h.registry.HealthCheck(r.Context())
	servers := make([]map[string]any, 0, len(results))
	for name, err := range results {
		entry := map[string]any{
			"name":   name,
			"healthy": err == nil,
		}
		if err != nil {
			entry["error"] = err.Error()
		}
		servers = append(servers, entry)
	}

	// Also add servers that are registered but not in health check results
	for _, name := range h.registry.ListMCPServers() {
		found := false
		for _, s := range servers {
			if s["name"] == name {
				found = true
				break
			}
		}
		if !found {
			servers = append(servers, map[string]any{
				"name":    name,
				"healthy": true,
			})
		}
	}

	writeEnvelope(w, http.StatusOK, "mcp_health", map[string]any{
		"servers": servers,
	})
}

// HealthCheckOne returns the health status of a specific MCP server.
func (h *MCPHealthHandler) HealthCheckOne(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id := strings.TrimSuffix(path, "/health")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return
	}

	server, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	if h.registry == nil || !h.registry.IsRegistered(server.Name) {
		writeEnvelope(w, http.StatusOK, "mcp_server_health", map[string]any{
			"server_id":   server.ID,
			"server_name": server.Name,
			"healthy":     false,
			"error":       "not registered in runtime",
		})
		return
	}

	results := h.registry.HealthCheck(r.Context())
	err, exists := results[server.Name]
	healthy := exists && err == nil

	entry := map[string]any{
		"server_id":   server.ID,
		"server_name": server.Name,
		"healthy":     healthy,
	}
	if err != nil {
		entry["error"] = err.Error()
	}

	writeEnvelope(w, http.StatusOK, "mcp_server_health", entry)
}
