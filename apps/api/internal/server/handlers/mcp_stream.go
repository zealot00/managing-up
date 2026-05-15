package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
)

// MCPStreamHandler handles SSE-based MCP tool invocation.
type MCPStreamHandler struct {
	registry    *executors.MCPRegistry
	repo        MCPInvokeRepository
	keyResolver APIKeyResolver
}

func NewMCPStreamHandler(registry *executors.MCPRegistry, repo MCPInvokeRepository, keyResolver APIKeyResolver) *MCPStreamHandler {
	return &MCPStreamHandler{registry: registry, repo: repo, keyResolver: keyResolver}
}

// InvokeStream handles SSE-based tool invocation with progress notifications.
func (h *MCPStreamHandler) InvokeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	var req InvokeMCPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.ServerID == "" || req.ToolName == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "server_id and tool_name are required.")
		return
	}

	// Resolve identity: try JWT user first, then API key
	ctx := r.Context()
	user := middleware.UserFromContext(ctx)
	userID := ""
	apiKeyID := ""

	if user != nil {
		userID = user.ID
	}

	rawAPIKey := getAPIKeyFromRequest(r)
	if rawAPIKey != "" && h.keyResolver != nil {
		dbKeyID, keyUserID, ok := h.keyResolver.ResolveAPIKey(rawAPIKey)
		if ok {
			apiKeyID = dbKeyID
			if userID == "" && keyUserID != "" {
				userID = keyUserID
			}
		}
	}

	// Permission check
	permitted, err := h.repo.CheckMCPPermission(req.ServerID, userID, apiKeyID, req.SkillID)
	if err != nil || !permitted {
		writeError(w, http.StatusForbidden, "PERMISSION_DENIED", "You do not have permission to invoke this MCP server.")
		return
	}

	// Get server config
	server, ok := h.repo.GetMCPServer(req.ServerID)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	if server.Status != "approved" || !server.IsEnabled {
		writeError(w, http.StatusForbidden, "SERVER_UNAVAILABLE", "MCP server is not approved or is disabled.")
		return
	}

	// Register client if needed
	if !h.registry.IsRegistered(server.Name) {
		switch server.TransportType {
		case "http", "https":
			headers := parseHeadersToMap(server.Headers)
			if err := h.registry.RegisterHTTP(ctx, server.Name, server.URL, headers); err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to register MCP client: %v", err))
				return
			}
		case "stdio":
			env := server.Env
			if env == nil {
				env = []string{}
			}
			if err := h.registry.RegisterStdio(ctx, server.Name, executors.MCPClientConfig{
				Command: server.Command,
				Args:    server.Args,
				Env:     env,
				Timeout: 30 * time.Second,
			}); err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("failed to register MCP client: %v", err))
				return
			}
		default:
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("unsupported transport: %s", server.TransportType))
			return
		}
	}

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "streaming not supported")
		return
	}

	sendEvent := func(event string, data any) {
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
		flusher.Flush()
	}

	// Execute tool call
	args := req.Parameters
	if args == nil {
		args = make(map[string]any)
	}

	tool, found := h.registry.GetToolByServer(server.Name, req.ToolName)
	if !found {
		sendEvent("error", map[string]any{"error": fmt.Sprintf("tool not found: %s on server %s", req.ToolName, server.Name)})
		sendEvent("done", map[string]any{})
		return
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		slog.Error("MCP tool invocation failed", "tool", req.ToolName, "server", server.Name, "error", err)
		sendEvent("error", map[string]any{"error": err.Error()})
		sendEvent("done", map[string]any{})
		return
	}

	// Increment use count
	_ = h.repo.IncrementMCPRouterCatalogUseCount(req.ServerID)

	sendEvent("result", result)
	sendEvent("done", map[string]any{})
}
