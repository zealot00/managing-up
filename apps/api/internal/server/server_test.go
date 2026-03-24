package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zealot/managing-up/apps/api/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status payload ok, got %q", body["status"])
	}
}

func TestMetaEndpoint(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	rec := httptest.NewRecorder()

	handleMeta(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data payload")
	}

	if data["service"] != "managing-up-api" {
		t.Fatalf("expected service name to be set, got %v", data["service"])
	}

	if data["runtime"] != "go" {
		t.Fatalf("expected runtime go, got %v", data["runtime"])
	}
}

func TestSkillsListEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCreateSkillEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"name":"test_skill","owner_team":"platform_team","risk_level":"medium"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestCreateExecutionSkillNotFound(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"skill_id":"skill_999","triggered_by":"operator","input":{"server_id":"srv-001"}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCreateExecutionEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"skill_id":"skill_001","triggered_by":"operator","input":{"server_id":"srv-001"}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestProcedureDraftsEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/procedure-drafts", nil)
	rec := httptest.NewRecorder()

	srv.handleProcedureDrafts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSkillVersionsEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skill-versions?skill_id=skill_001", nil)
	rec := httptest.NewRecorder()

	srv.handleSkillVersions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestApprovalsEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/approvals", nil)
	rec := httptest.NewRecorder()

	srv.handleApprovals(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestExecutionApprovalEndpoint(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"approver":"ops_manager","decision":"approved","note":"safe to continue"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/exec_002/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutionByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCreateSkillValidationMissingName(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"owner_team":"platform_team","risk_level":"medium"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateSkillValidationMissingOwnerTeam(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"name":"test_skill","risk_level":"medium"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateSkillValidationInvalidRiskLevel(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"name":"test_skill","owner_team":"platform_team","risk_level":"critical"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateSkillValidationInvalidContentType(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"name":"test_skill","owner_team":"platform_team","risk_level":"medium"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestCreateSkillValidationInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"name":"test_skill","owner_team":}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleSkills(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateExecutionValidationMissingSkillID(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"triggered_by":"operator","input":{"server_id":"srv-001"}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateExecutionValidationMissingTriggeredBy(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"skill_id":"skill_001","input":{"server_id":"srv-001"}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateExecutionValidationInvalidContentType(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"skill_id":"skill_001","triggered_by":"operator","input":{"server_id":"srv-001"}}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestCreateExecutionValidationInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"skill_id":}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveExecutionValidationMissingApprover(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"decision":"approved","note":"safe to continue"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/exec_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutionByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveExecutionValidationInvalidDecision(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"approver":"ops_manager","decision":"maybe","note":"not sure"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/exec_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutionByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveExecutionValidationInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"approver":}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/exec_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutionByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveExecutionNotFound(t *testing.T) {
	t.Parallel()

	srv := New(config.Config{Port: "8080"})
	body := []byte(`{"approver":"ops_manager","decision":"approved","note":"safe to continue"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/nonexistent/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleExecutionByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Dashboard POST", http.MethodPost, "/api/v1/dashboard"},
		{"ProcedureDrafts POST", http.MethodPost, "/api/v1/procedure-drafts"},
		{"SkillVersions POST", http.MethodPost, "/api/v1/skill-versions"},
		{"Approvals POST", http.MethodPost, "/api/v1/approvals"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := New(config.Config{Port: "8080"})
			body := []byte(`{}`)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.handleDashboard(rec, req)
			if tt.path == "/api/v1/procedure-drafts" {
				srv.handleProcedureDrafts(rec, req)
			} else if tt.path == "/api/v1/skill-versions" {
				srv.handleSkillVersions(rec, req)
			} else if tt.path == "/api/v1/approvals" {
				srv.handleApprovals(rec, req)
			}

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
			}
		})
	}
}
