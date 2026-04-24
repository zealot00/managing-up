package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

type PoliciesHandler struct {
	repo interface {
		ListPolicyVersions() ([]models.PolicyVersion, error)
		GetPolicyVersion(name string) (models.PolicyVersion, bool)
		CreatePolicyVersion(pv models.PolicyVersion) (models.PolicyVersion, error)
		UpdatePolicyVersion(pv models.PolicyVersion) error
		DeletePolicyVersion(id string) error
	}
}

func NewPoliciesHandler(repo interface {
	ListPolicyVersions() ([]models.PolicyVersion, error)
	GetPolicyVersion(name string) (models.PolicyVersion, bool)
	CreatePolicyVersion(pv models.PolicyVersion) (models.PolicyVersion, error)
	UpdatePolicyVersion(pv models.PolicyVersion) error
	DeletePolicyVersion(id string) error
}) *PoliciesHandler {
	return &PoliciesHandler{repo: repo}
}

type PolicyVersionDTO struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	IsActive    bool            `json:"is_active"`
	Rules       []PolicyRuleDTO `json:"rules"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type PolicyRuleDTO struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Condition string `json:"condition"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	Priority  int    `json:"priority"`
	IsActive  bool   `json:"is_active"`
}

type CreatePolicyVersionRequest struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	Rules       []PolicyRuleDTO `json:"rules"`
}

func (h *PoliciesHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	versions, err := h.repo.ListPolicyVersions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	items := make([]PolicyVersionDTO, 0, len(versions))
	for _, v := range versions {
		items = append(items, toPolicyVersionDTO(v))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (h *PoliciesHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractPathID(r.URL.Path, "policies")
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "policy id is required")
		return
	}

	version, ok := h.repo.GetPolicyVersion(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "policy not found")
		return
	}

	writeJSON(w, http.StatusOK, toPolicyVersionDTO(version))
}

func (h *PoliciesHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	var req CreatePolicyVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "NAME_REQUIRED", "name is required")
		return
	}
	if req.Version == "" {
		req.Version = "v1"
	}

	now := time.Now()
	pv := &models.PolicyVersion{
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rules := make([]models.PolicyRule, 0, len(req.Rules))
	for _, r := range req.Rules {
		rules = append(rules, models.PolicyRule{
			ID:        r.ID,
			Version:   r.Version,
			Condition: r.Condition,
			Action:    r.Action,
			Reason:    r.Reason,
			Priority:  r.Priority,
			IsActive:  r.IsActive,
		})
	}
	pv.Rules = rules

	result, err := h.repo.CreatePolicyVersion(*pv)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": result.ID, "name": result.Name})
}

func (h *PoliciesHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "")
		return
	}

	id := extractPathID(r.URL.Path, "policies")
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "policy id is required")
		return
	}

	var req CreatePolicyVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	pv, ok := h.repo.GetPolicyVersion(id)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "policy not found")
		return
	}

	pv.Description = req.Description
	pv.IsDefault = req.IsDefault
	pv.UpdatedAt = time.Now()

	rules := make([]models.PolicyRule, 0, len(req.Rules))
	for _, r := range req.Rules {
		rules = append(rules, models.PolicyRule{
			ID:        r.ID,
			Version:   r.Version,
			Condition: r.Condition,
			Action:    r.Action,
			Reason:    r.Reason,
			Priority:  r.Priority,
			IsActive:  r.IsActive,
		})
	}
	pv.Rules = rules

	if err := h.repo.UpdatePolicyVersion(pv); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func toPolicyVersionDTO(pv models.PolicyVersion) PolicyVersionDTO {
	rules := make([]PolicyRuleDTO, 0, len(pv.Rules))
	for _, r := range pv.Rules {
		rules = append(rules, PolicyRuleDTO{
			ID:        r.ID,
			Version:   r.Version,
			Condition: r.Condition,
			Action:    r.Action,
			Reason:    r.Reason,
			Priority:  r.Priority,
			IsActive:  r.IsActive,
		})
	}

	return PolicyVersionDTO{
		ID:          pv.ID,
		Name:        pv.Name,
		Version:     pv.Version,
		Description: pv.Description,
		IsDefault:   pv.IsDefault,
		IsActive:    pv.IsActive,
		Rules:       rules,
		CreatedAt:   pv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   pv.UpdatedAt.Format(time.RFC3339),
	}
}

func extractPathID(path, prefix string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == prefix && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}