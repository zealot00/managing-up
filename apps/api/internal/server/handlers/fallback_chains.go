package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/gateway"
	"github.com/zealot/managing-up/apps/api/internal/llm"
	"github.com/zealot/managing-up/apps/api/internal/server/middleware"
)

// FallbackChainDTO is the data transfer object for fallback chains.
type FallbackChainDTO struct {
	ID        string                `json:"id"`
	Model     string                `json:"model"`
	IsEnabled bool                  `json:"is_enabled"`
	Targets   []FallbackTargetDTO   `json:"targets"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// FallbackTargetDTO is the data transfer object for fallback targets.
type FallbackTargetDTO struct {
	ID        string `json:"id"`
	ChainID   string `json:"chain_id"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Weight    int    `json:"weight"`
	Priority  int    `json:"priority"`
	IsEnabled bool   `json:"is_enabled"`
}

// FallbackChainRepo defines the repository interface needed by the handler.
type FallbackChainRepo interface {
	ListFallbackChains() ([]FallbackChainDTO, error)
	GetFallbackChain(id string) (FallbackChainDTO, bool, error)
	CreateFallbackChain(chain FallbackChainDTO) (FallbackChainDTO, error)
	UpdateFallbackChain(chain FallbackChainDTO) (FallbackChainDTO, error)
	DeleteFallbackChain(id string) error
}

// FallbackChainReloader hot-reloads fallback chains into the router.
type FallbackChainReloader interface {
	SetFallbackChains(chains map[llm.Model][]gateway.FallbackTarget)
}

// FallbackChainHandler handles CRUD operations for fallback chain configuration.
type FallbackChainHandler struct {
	repo     FallbackChainRepo
	reloader FallbackChainReloader
	authMW   *middleware.AuthMiddleware
}

// NewFallbackChainHandler creates a new handler.
func NewFallbackChainHandler(repo FallbackChainRepo, reloader FallbackChainReloader, authMW *middleware.AuthMiddleware) *FallbackChainHandler {
	return &FallbackChainHandler{repo: repo, reloader: reloader, authMW: authMW}
}

// RegisterRoutes registers fallback chain endpoints on the mux with auth protection.
func (h *FallbackChainHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/admin/fallback-chains", h.authMW.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.list(w, r)
		case http.MethodPost:
			h.create(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})))

	mux.Handle("/api/v1/admin/fallback-chains/", h.authMW.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path: /api/v1/admin/fallback-chains/{id}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/fallback-chains/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID is required")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.get(w, r, id)
		case http.MethodPut:
			h.update(w, r, id)
		case http.MethodDelete:
			h.delete(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})))
}

func (h *FallbackChainHandler) list(w http.ResponseWriter, r *http.Request) {
	chains, err := h.repo.ListFallbackChains()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to list fallback chains: %v", err))
		return
	}
	if chains == nil {
		chains = []FallbackChainDTO{}
	}
	writeEnvelope(w, http.StatusOK, "fallback_chains", map[string]any{"items": chains})
}

func (h *FallbackChainHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	chain, found, err := h.repo.GetFallbackChain(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to get fallback chain: %v", err))
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Fallback chain not found")
		return
	}
	writeEnvelope(w, http.StatusOK, "fallback_chain", chain)
}

type createFallbackChainRequest struct {
	Model     string              `json:"model"`
	IsEnabled *bool               `json:"is_enabled"`
	Targets   []FallbackTargetDTO `json:"targets"`
}

func (h *FallbackChainHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createFallbackChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "model is required")
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	chain := FallbackChainDTO{
		Model:     req.Model,
		IsEnabled: isEnabled,
		Targets:   req.Targets,
	}

	created, err := h.repo.CreateFallbackChain(chain)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to create fallback chain: %v", err))
		return
	}

	h.reloadChains()

	writeEnvelope(w, http.StatusCreated, "fallback_chain_created", created)
}

type updateFallbackChainRequest struct {
	Model     string              `json:"model"`
	IsEnabled *bool               `json:"is_enabled"`
	Targets   []FallbackTargetDTO `json:"targets"`
}

func (h *FallbackChainHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var req updateFallbackChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	existing, found, err := h.repo.GetFallbackChain(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to get fallback chain: %v", err))
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Fallback chain not found")
		return
	}

	if req.Model != "" {
		existing.Model = req.Model
	}
	if req.IsEnabled != nil {
		existing.IsEnabled = *req.IsEnabled
	}
	if req.Targets != nil {
		existing.Targets = req.Targets
	}

	updated, err := h.repo.UpdateFallbackChain(existing)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to update fallback chain: %v", err))
		return
	}

	h.reloadChains()

	writeEnvelope(w, http.StatusOK, "fallback_chain_updated", updated)
}

func (h *FallbackChainHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.repo.DeleteFallbackChain(id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to delete fallback chain: %v", err))
		return
	}

	h.reloadChains()

	writeEnvelope(w, http.StatusOK, "fallback_chain_deleted", map[string]any{"deleted": true})
}

// reloadChains reads all chains from the DB and pushes them into the router.
func (h *FallbackChainHandler) reloadChains() {
	if h.reloader == nil {
		return
	}

	chains, err := h.repo.ListFallbackChains()
	if err != nil {
		return
	}

	routerChains := make(map[llm.Model][]gateway.FallbackTarget)
	for _, c := range chains {
		if !c.IsEnabled {
			continue
		}
		var targets []gateway.FallbackTarget
		for _, t := range c.Targets {
			if t.IsEnabled {
				targets = append(targets, gateway.FallbackTarget{
					Provider: llm.Provider(t.Provider),
					Model:    llm.Model(t.Model),
				})
			}
		}
		if len(targets) > 0 {
			routerChains[llm.Model(c.Model)] = targets
		}
	}

	h.reloader.SetFallbackChains(routerChains)
}
