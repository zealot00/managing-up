package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
	"github.com/zealot/managing-up/apps/api/internal/service"
)

type MCPServersRepository interface {
	GetMCPServer(id string) (MCPServerDTO, bool)
	UpdateMCPServer(server MCPServerDTO) error
}

type MCPServerDTO struct {
	ID              string
	Name            string
	Description     string
	TransportType   string
	Command         string
	Args            []string
	Env             []string
	URL             string
	Headers         []string
	Status          string
	RejectionReason string
	ApprovedBy      string
	ApprovedAt      *time.Time
	IsEnabled       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type MCPServersHandler struct {
	repo         MCPServersRepository
	mcpRouterSvc *service.MCPRouterService
	registry     *executors.MCPRegistry
}

func NewMCPServersHandler(repo MCPServersRepository, mcpRouterSvc *service.MCPRouterService, registry *executors.MCPRegistry) *MCPServersHandler {
	return &MCPServersHandler{
		repo:         repo,
		mcpRouterSvc: mcpRouterSvc,
		registry:     registry,
	}
}

type ApproveMCPServerRequest struct {
	Decision string `json:"decision"`
	Approver string `json:"approver"`
	Note     string `json:"note,omitempty"`
}

func (h *MCPServersHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id := strings.TrimSuffix(path, "/approve")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json.")
		return
	}

	var req ApproveMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.Decision != "approved" && req.Decision != "rejected" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "decision must be 'approved' or 'rejected'.")
		return
	}

	ctx := r.Context()
	mcpServer, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	if req.Decision == "approved" {
		validationResult := executors.ValidateMCPServer(ctx, executors.MCPServerConfig{
			TransportType: mcpServer.TransportType,
			Command:       mcpServer.Command,
			Args:          mcpServer.Args,
			Env:           mcpServer.Env,
			URL:           mcpServer.URL,
			Headers:       mcpServer.Headers,
		})
		if !validationResult.Valid {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED",
				fmt.Sprintf("MCP server validation failed: %s", validationResult.Error))
			return
		}
	}

	now := time.Now()
	mcpServer.Status = req.Decision
	mcpServer.ApprovedBy = req.Approver
	mcpServer.ApprovedAt = &now

	if req.Decision == "rejected" && req.Note != "" {
		mcpServer.RejectionReason = req.Note
	}

	if err := h.repo.UpdateMCPServer(mcpServer); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update MCP server.")
		return
	}

	if req.Decision == "approved" {
		user := middleware.UserFromContext(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.ID
		}
		syncServer := service.MCPServer{
			ID:            mcpServer.ID,
			Name:          mcpServer.Name,
			TrustScore:    0.5,
			TransportType: mcpServer.TransportType,
			URL:           mcpServer.URL,
		}
		if err := h.mcpRouterSvc.SyncFromMCPServer(ctx, syncServer, approvedBy); err != nil {
			log.Printf("failed to sync to router catalog: %v", err)
		}

		// Register the server into the runtime MCPRegistry so tools are discoverable immediately
		if h.registry != nil && !h.registry.IsRegistered(mcpServer.Name) {
			go h.registerToRuntime(mcpServer)
		}
	}

	writeEnvelope(w, http.StatusOK, "req_mcp_approve", mcpServer)
}

// ListTools returns the tools available on a specific MCP server.
func (h *MCPServersHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id := strings.TrimSuffix(path, "/tools")
	if id == "" || id == path {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server ID is required.")
		return
	}

	mcpServer, ok := h.repo.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found.")
		return
	}

	if h.registry == nil {
		writeEnvelope(w, http.StatusOK, "mcp_server_tools", map[string]any{
			"server_id":   mcpServer.ID,
			"server_name": mcpServer.Name,
			"tools":       []any{},
		})
		return
	}

	// If server is approved but not yet in runtime registry, register it on the fly
	if !h.registry.IsRegistered(mcpServer.Name) && mcpServer.Status == "approved" && mcpServer.IsEnabled {
		h.registerToRuntimeSync(r.Context(), mcpServer)
	}

	mcpTools := h.registry.ListToolsByServer(mcpServer.Name)
	tools := make([]map[string]any, 0, len(mcpTools))
	for _, t := range mcpTools {
		toolInfo := map[string]any{
			"name":        t.Name,
			"description": t.Description,
		}
		if t.InputSchema.Type != "" || len(t.InputSchema.Properties) > 0 {
			toolInfo["inputSchema"] = t.InputSchema
		}
		tools = append(tools, toolInfo)
	}

	writeEnvelope(w, http.StatusOK, "mcp_server_tools", map[string]any{
		"server_id":   mcpServer.ID,
		"server_name": mcpServer.Name,
		"tools":       tools,
	})
}

// ListAllTools returns all tools from all registered MCP servers.
func (h *MCPServersHandler) ListAllTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if h.registry == nil {
		writeEnvelope(w, http.StatusOK, "mcp_all_tools", map[string]any{
			"tools": []any{},
		})
		return
	}

	allTools := h.registry.ListTools()
	tools := make([]map[string]any, 0, len(allTools))
	for _, t := range allTools {
		mcpTool, ok := t.(*executors.MCPTool)
		if !ok {
			continue
		}
		info := executors.MCPToolInfoFromMCP(mcpTool.ServerName(), mcpTool.MCPToolDef())
		toolInfo := map[string]any{
			"server_name": info.ServerName,
			"name":        info.Name,
			"description": info.Description,
		}
		if info.InputSchema != nil {
			toolInfo["inputSchema"] = info.InputSchema
		}
		tools = append(tools, toolInfo)
	}

	writeEnvelope(w, http.StatusOK, "mcp_all_tools", map[string]any{
		"tools": tools,
	})
}

// registerToRuntime registers an approved MCP server into the in-memory registry.
// Called in a goroutine from Approve to avoid blocking the HTTP response.
func (h *MCPServersHandler) registerToRuntime(server MCPServerDTO) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	h.doRegisterToRuntime(ctx, server)
}

// registerToRuntimeSync registers synchronously (used by ListTools when server not yet in registry).
func (h *MCPServersHandler) registerToRuntimeSync(ctx context.Context, server MCPServerDTO) {
	h.doRegisterToRuntime(ctx, server)
}

func (h *MCPServersHandler) doRegisterToRuntime(ctx context.Context, server MCPServerDTO) {
	switch server.TransportType {
	case "http", "https":
		headers := parseHeadersToMap(server.Headers)
		if err := h.registry.RegisterHTTP(ctx, server.Name, server.URL, headers); err != nil {
			log.Printf("failed to register MCP server %s to runtime registry: %v", server.Name, err)
		} else {
			log.Printf("registered MCP server %s to runtime registry", server.Name)
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
			log.Printf("failed to register MCP server %s to runtime registry: %v", server.Name, err)
		} else {
			log.Printf("registered MCP server %s to runtime registry", server.Name)
		}
	}
}
