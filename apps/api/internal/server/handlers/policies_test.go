package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

func TestPoliciesHandler_ListPolicies(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{
		versions: []models.PolicyVersion{
			{
				ID:          "policy_001",
				Name:        "strict_policy",
				Version:     "v1",
				Description: "Strict release policy",
				IsDefault:   true,
				IsActive:    true,
				Rules:       []models.PolicyRule{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "policy_002",
				Name:        "lenient_policy",
				Version:     "v2",
				Description: "Lenient policy",
				IsDefault:   false,
				IsActive:    true,
				Rules:       []models.PolicyRule{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies", nil)
	rec := httptest.NewRecorder()

	handler.ListPolicies(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	items, ok := body["items"].([]any)
	if !ok {
		t.Fatalf("expected items array in response")
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(items))
	}
}

func TestPoliciesHandler_ListPolicies_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies", nil)
	rec := httptest.NewRecorder()

	handler.ListPolicies(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestPoliciesHandler_GetPolicy(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{
		versions: []models.PolicyVersion{
			{
				ID:          "policy_001",
				Name:        "strict_policy",
				Version:     "v1",
				Description: "Strict release policy",
				IsDefault:   true,
				IsActive:    true,
				Rules: []models.PolicyRule{
					{
						ID:        "rule_001",
						Version:   "v1",
						Condition: "task_type == 'code_gen'",
						Action:    "allow",
						Reason:    "code gen is allowed",
						Priority:  1,
						IsActive:  true,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/policy_001", nil)
	rec := httptest.NewRecorder()

	handler.GetPolicy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	if body["id"] != "policy_001" {
		t.Fatalf("expected policy id policy_001, got %v", body["id"])
	}

	if body["name"] != "strict_policy" {
		t.Fatalf("expected name strict_policy, got %v", body["name"])
	}
}

func TestPoliciesHandler_GetPolicy_NotFound(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.GetPolicy(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestPoliciesHandler_GetPolicy_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies/policy_001", nil)
	rec := httptest.NewRecorder()

	handler.GetPolicy(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestPoliciesHandler_CreatePolicy(t *testing.T) {
	t.Parallel()

	repo := &mockPoliciesRepo{}
	handler := NewPoliciesHandler(repo)

	body := []byte(`{
		"name": "new_policy",
		"version": "v1",
		"description": "A new policy",
		"is_default": true,
		"rules": [
			{
				"id": "rule_001",
				"version": "v1",
				"condition": "task_type == 'test'",
				"action": "allow",
				"reason": "testing allowed",
				"priority": 1,
				"is_active": true
			}
		]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreatePolicy(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if !repo.created {
		t.Fatalf("expected CreatePolicyVersion to be called")
	}
}

func TestPoliciesHandler_CreatePolicy_ValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		errMsg     string
		statusCode int
	}{
		{
			name:       "missing name",
			body:       `{"version": "v1"}`,
			errMsg:     "NAME_REQUIRED",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{"name":}`,
			errMsg:     "INVALID_BODY",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewPoliciesHandler(&mockPoliciesRepo{})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/policies", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreatePolicy(rec, req)

			if rec.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, rec.Code)
			}
		})
	}
}

func TestPoliciesHandler_CreatePolicy_DefaultVersion(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	body := []byte(`{"name": "test_policy"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreatePolicy(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestPoliciesHandler_UpdatePolicy(t *testing.T) {
	t.Parallel()

	repo := &mockPoliciesRepo{
		versions: []models.PolicyVersion{
			{
				ID:          "policy_001",
				Name:        "old_name",
				Version:     "v1",
				Description: "Old description",
				IsDefault:   false,
				IsActive:    true,
				Rules:       []models.PolicyRule{},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}
	handler := NewPoliciesHandler(repo)

	body := []byte(`{
		"description": "New description",
		"is_default": true
	}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/policies/policy_001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdatePolicy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !repo.updated {
		t.Fatalf("expected UpdatePolicyVersion to be called")
	}
}

func TestPoliciesHandler_UpdatePolicy_NotFound(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	body := []byte(`{"description": "New description"}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/policies/nonexistent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdatePolicy(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestPoliciesHandler_UpdatePolicy_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/policy_001", nil)
	rec := httptest.NewRecorder()

	handler.UpdatePolicy(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestPoliciesHandler_UpdatePolicy_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewPoliciesHandler(&mockPoliciesRepo{})

	body := []byte(`{"description":}`)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/policies/policy_001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdatePolicy(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestToPolicyVersionDTO(t *testing.T) {
	t.Parallel()

	now := time.Now()
	pv := models.PolicyVersion{
		ID:          "policy_001",
		Name:        "test_policy",
		Version:     "v1",
		Description: "Test description",
		IsDefault:   true,
		IsActive:    true,
		Rules: []models.PolicyRule{
			{
				ID:        "rule_001",
				Version:   "v1",
				Condition: "true",
				Action:    "allow",
				Reason:    "test",
				Priority:  1,
				IsActive:  true,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	dto := toPolicyVersionDTO(pv)

	if dto.ID != pv.ID {
		t.Fatalf("expected ID %s, got %s", pv.ID, dto.ID)
	}
	if dto.Name != pv.Name {
		t.Fatalf("expected Name %s, got %s", pv.Name, dto.Name)
	}
	if dto.Version != pv.Version {
		t.Fatalf("expected Version %s, got %s", pv.Version, dto.Version)
	}
	if dto.IsDefault != pv.IsDefault {
		t.Fatalf("expected IsDefault %v, got %v", pv.IsDefault, dto.IsDefault)
	}
	if len(dto.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(dto.Rules))
	}
}

func TestExtractPathID(t *testing.T) {
	tests := []struct {
		path     string
		prefix   string
		expected string
	}{
		{"/api/v1/policies/policy_001", "policies", "policy_001"},
		{"/policies/policy_002", "policies", "policy_002"},
		{"policies/test", "policies", "test"},
		{"/api/v1/other/value", "policies", ""},
		{"/api/v1/policies/", "policies", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractPathID(tt.path, tt.prefix)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

type mockPoliciesRepo struct {
	versions []models.PolicyVersion
	created  bool
	updated  bool
}

func (m *mockPoliciesRepo) ListPolicyVersions() ([]models.PolicyVersion, error) {
	return m.versions, nil
}

func (m *mockPoliciesRepo) GetPolicyVersion(name string) (models.PolicyVersion, bool) {
	for _, v := range m.versions {
		if v.ID == name {
			return v, true
		}
	}
	return models.PolicyVersion{}, false
}

func (m *mockPoliciesRepo) CreatePolicyVersion(pv models.PolicyVersion) (models.PolicyVersion, error) {
	m.created = true
	pv.ID = "new_policy_id"
	return pv, nil
}

func (m *mockPoliciesRepo) UpdatePolicyVersion(pv models.PolicyVersion) error {
	m.updated = true
	return nil
}

func (m *mockPoliciesRepo) DeletePolicyVersion(id string) error {
	return nil
}