package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
)

func getAPIKeyFromRequest(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if apiKey := r.Header.Get("x-api-key"); apiKey != "" {
		return apiKey
	}
	return ""
}

func parseHeadersToMap(headers []string) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		if parts := strings.SplitN(h, ":", 2); len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

type MCPInvokeHandler struct {
	repo         MCPInvokeRepository
	mcpRouterSvc MCPInvokerService
}

type MCPInvokeRepository interface {
	CheckMCPPermission(mcpServerID, userID, apiKeyID, skillID string) (bool, error)
	IncrementMCPRouterCatalogUseCount(serverID string) error
	GetMCPServer(id string) (MCPServerDTO, bool)
}

type MCPServerConfig struct {
	ServerID      string
	Name         string
	TransportType string
	URL          string
	Command      string
	Args         []string
	Env          []string
	Headers      map[string]string
}

type MCPInvokerService interface {
	InvokeTool(ctx context.Context, config MCPServerConfig, toolName string, params map[string]interface{}) (*MCPInvokeResult, error)
}

type MCPInvokeResult struct {
	Success bool                   `json:"success"`
	Output  map[string]interface{}   `json:"output,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type InvokeMCPRequest struct {
	ServerID   string                 `json:"server_id"`
	ToolName  string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	APIKeyID  string                 `json:"api_key_id,omitempty"`
	SkillID   string                 `json:"skill_id,omitempty"`
}

func NewMCPInvokeHandler(repo MCPInvokeRepository, mcpRouterSvc MCPInvokerService) *MCPInvokeHandler {
	return &MCPInvokeHandler{repo: repo, mcpRouterSvc: mcpRouterSvc}
}

func (h *MCPInvokeHandler) Invoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
		return
	}

	var req InvokeMCPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.ServerID == "" {
		writeError(w, http.StatusBadRequest, "SERVER_ID_REQUIRED", "server_id is required")
		return
	}
	if req.ToolName == "" {
		writeError(w, http.StatusBadRequest, "TOOL_NAME_REQUIRED", "tool_name is required")
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user != nil {
		req.UserID = user.ID
	}

	apiKeyID := getAPIKeyFromRequest(r)
	if apiKeyID != "" {
		req.APIKeyID = apiKeyID
	}

	hasPermission, err := h.repo.CheckMCPPermission(req.ServerID, req.UserID, req.APIKeyID, req.SkillID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PERMISSION_CHECK_FAILED", "failed to check permission")
		return
	}
	if !hasPermission {
		writeError(w, http.StatusForbidden, "PERMISSION_DENIED", "user/api_key does not have permission to invoke this MCP server")
		return
	}

	server, ok := h.repo.GetMCPServer(req.ServerID)
	if !ok {
		writeError(w, http.StatusNotFound, "SERVER_NOT_FOUND", "MCP server not found")
		return
	}

	config := MCPServerConfig{
		ServerID:      server.ID,
		Name:          server.Name,
		TransportType: server.TransportType,
		URL:           server.URL,
		Command:       server.Command,
		Args:          server.Args,
		Env:           server.Env,
		Headers:       parseHeadersToMap(server.Headers),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := h.mcpRouterSvc.InvokeTool(ctx, config, req.ToolName, req.Parameters)
	if err != nil {
		h.repo.IncrementMCPRouterCatalogUseCount(req.ServerID)
		writeError(w, http.StatusInternalServerError, "INVOKE_FAILED", err.Error())
		return
	}

	h.repo.IncrementMCPRouterCatalogUseCount(req.ServerID)

	if result.Success {
		writeEnvelope(w, http.StatusOK, "mcp_invoke", map[string]any{
			"success": true,
			"output":  result.Output,
		})
	} else {
		writeEnvelope(w, http.StatusOK, "mcp_invoke", map[string]any{
			"success": false,
			"error":   result.Error,
		})
	}
}

type GrantMCPHandler struct {
	repo MCPGrantRepository
}

type MCPGrantRepository interface {
	CreateMCPServerPermission(p MCPServerPermission) (MCPServerPermission, error)
	ListMCPServerPermissions(mcpServerID string) ([]MCPServerPermission, error)
	GetMCPServer(id string) (MCPServerDTO, bool)
}

type MCPServerPermission struct {
	ID              string     `json:"id"`
	MCPServerID    string     `json:"mcp_server_id"`
	UserID         string     `json:"user_id,omitempty"`
	APIKeyID       string     `json:"api_key_id,omitempty"`
	SkillID        string     `json:"skill_id,omitempty"`
	PermissionType string     `json:"permission_type"`
	IsGranted      bool       `json:"is_granted"`
	GrantedBy      string     `json:"granted_by,omitempty"`
	GrantedAt      time.Time  `json:"granted_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type GrantMCPRequest struct {
	MCPServerID    string     `json:"mcp_server_id"`
	UserID         string    `json:"user_id,omitempty"`
	APIKeyID       string    `json:"api_key_id,omitempty"`
	SkillID        string    `json:"skill_id,omitempty"`
	PermissionType string    `json:"permission_type"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

func NewGrantMCPHandler(repo MCPGrantRepository) *GrantMCPHandler {
	return &GrantMCPHandler{repo: repo}
}

func (h *GrantMCPHandler) Grant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	if !isJSONRequest(r) {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
		return
	}

	var req GrantMCPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if req.MCPServerID == "" {
		writeError(w, http.StatusBadRequest, "SERVER_ID_REQUIRED", "mcp_server_id is required")
		return
	}

	if req.UserID == "" && req.APIKeyID == "" && req.SkillID == "" {
		writeError(w, http.StatusBadRequest, "TARGET_REQUIRED", "user_id, api_key_id, or skill_id is required")
		return
	}

	_, ok := h.repo.GetMCPServer(req.MCPServerID)
	if !ok {
		writeError(w, http.StatusNotFound, "SERVER_NOT_FOUND", "MCP server not found")
		return
	}

	user := middleware.UserFromContext(r.Context())
	grantedBy := ""
	if user != nil {
		grantedBy = user.ID
	}

	perm := MCPServerPermission{
		MCPServerID:    req.MCPServerID,
		UserID:         req.UserID,
		APIKeyID:       req.APIKeyID,
		SkillID:        req.SkillID,
		PermissionType: req.PermissionType,
		IsGranted:      true,
		GrantedBy:      grantedBy,
		GrantedAt:      time.Now(),
		ExpiresAt:      req.ExpiresAt,
	}

	result, err := h.repo.CreateMCPServerPermission(perm)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GRANT_FAILED", "failed to create permission")
		return
	}

	writeEnvelope(w, http.StatusCreated, "mcp_grant", result)
}

func (h *GrantMCPHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method)
		return
	}

	serverID := r.URL.Query().Get("mcp_server_id")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "SERVER_ID_REQUIRED", "mcp_server_id is required")
		return
	}

	perms, err := h.repo.ListMCPServerPermissions(serverID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeEnvelope(w, http.StatusOK, "mcp_permissions", map[string]any{
		"items": perms,
	})
}