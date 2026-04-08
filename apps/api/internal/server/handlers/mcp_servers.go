package handlers

import (
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
}

func NewMCPServersHandler(repo MCPServersRepository, mcpRouterSvc *service.MCPRouterService) *MCPServersHandler {
	return &MCPServersHandler{
		repo:         repo,
		mcpRouterSvc: mcpRouterSvc,
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
	}

	writeEnvelope(w, http.StatusOK, "req_mcp_approve", mcpServer)
}
