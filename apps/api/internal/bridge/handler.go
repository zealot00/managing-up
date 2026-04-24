package bridge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zealot/managing-up/apps/api/internal/bridge/parser"
	"github.com/zealot/managing-up/apps/api/internal/bridge/template"
)

type BridgeAdapterConfig struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	AdapterType string         `json:"adapter_type"`
	Config      map[string]any `json:"config"`
	Tools       []map[string]any `json:"tools"`
	Enabled     bool           `json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type BridgeAdapterConfigRepository interface {
	ListBridgeAdapterConfigs() []BridgeAdapterConfig
	GetBridgeAdapterConfig(id string) (BridgeAdapterConfig, bool)
	CreateBridgeAdapterConfig(config BridgeAdapterConfig) (BridgeAdapterConfig, error)
	UpdateBridgeAdapterConfig(config BridgeAdapterConfig) error
	DeleteBridgeAdapterConfig(id string) error
}

type repoAdapter struct {
	repo BridgeAdapterConfigRepository
}

type Handler struct {
	repo BridgeAdapterConfigRepository
}

func NewHandler(repo BridgeAdapterConfigRepository) *Handler {
	return &Handler{repo: repo}
}

type CreateAdapterRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OpenAPISpec string                 `json:"openapi_spec"`
	Mappings    []template.ToolMapping `json:"mappings"`
	Options     template.AdapterOptions `json:"options"`
}

type AdapterResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ToolsCount  int    `json:"tools_count"`
	CreatedAt   string `json:"created_at"`
}

func (h *Handler) CreateAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	var req CreateAdapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "NAME_REQUIRED", "name is required")
		return
	}

	openAPISpec := req.OpenAPISpec
	spec, err := parser.ParseOpenAPI([]byte(openAPISpec))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_OPENAPI_SPEC", err.Error())
		return
	}

	endpoints := spec.ParseEndpoints()

	adapterTemplate := &template.AdapterTemplate{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OpenAPISpec: openAPISpec,
		Mappings:    req.Mappings,
		Options:     req.Options,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	generatedTools := adapterTemplate.GenerateTools(endpoints)
	toolsJSON, _ := json.Marshal(generatedTools)

	var tools []map[string]any
	json.Unmarshal(toolsJSON, &tools)

	config := BridgeAdapterConfig{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		AdapterType: "openapi",
		Config: map[string]any{
			"openapi_spec": openAPISpec,
			"options":      req.Options,
		},
		Tools:     tools,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	h.repo.CreateBridgeAdapterConfig(config)

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":   config.ID,
		"name": config.Name,
	})
}

func (h *Handler) ListAdapters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	configs := h.repo.ListBridgeAdapterConfigs()
	var items []AdapterResponse
	for _, c := range configs {
		items = append(items, AdapterResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			ToolsCount:  len(c.Tools),
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (h *Handler) GetAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "adapter id is required")
		return
	}

	config, ok := h.repo.GetBridgeAdapterConfig(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "adapter not found")
		return
	}

	writeJSON(w, http.StatusOK, config)
}

func (h *Handler) DeleteAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "adapter id is required")
		return
	}

	if _, ok := h.repo.GetBridgeAdapterConfig(id); !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "adapter not found")
		return
	}

	h.repo.DeleteBridgeAdapterConfig(id)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) TestAdapter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "adapter id is required")
		return
	}

	_, ok := h.repo.GetBridgeAdapterConfig(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "adapter not found")
		return
	}

	var req struct {
		ToolName string         `json:"tool_name"`
		Input    map[string]any `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tool_name":  req.ToolName,
		"input":      req.Input,
		"adapter_id": id,
		"status":     "ready to test",
	})
}

func (h *Handler) GetAdapterTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractID(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "adapter id is required")
		return
	}

	config, ok := h.repo.GetBridgeAdapterConfig(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "adapter not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"adapter_id": config.ID,
		"tools":      config.Tools,
	})
}

func extractID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[1] == "adapters" {
		return parts[2]
	}
	return ""
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/adapters", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListAdapters(w, r)
		case http.MethodPost:
			h.CreateAdapter(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		}
	})

	mux.HandleFunc("/api/v1/adapters/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/tools") {
			h.GetAdapterTools(w, r)
			return
		}
		if strings.HasSuffix(path, "/test") {
			h.TestAdapter(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetAdapter(w, r)
		case http.MethodDelete:
			h.DeleteAdapter(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		}
	})
}