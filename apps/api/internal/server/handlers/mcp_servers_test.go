package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMCPServersHandler_Approve_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewMCPServersHandler(&mockMCPServersRepo{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers/server_001/approve", nil)
	rec := httptest.NewRecorder()

	handler.Approve(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestMCPServersHandler_Approve_InvalidContentType(t *testing.T) {
	t.Parallel()

	handler := NewMCPServersHandler(&mockMCPServersRepo{}, nil)

	body := []byte(`{"decision": "approved", "approver": "admin"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers/server_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.Approve(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestMCPServersHandler_Approve_InvalidDecision(t *testing.T) {
	t.Parallel()

	handler := NewMCPServersHandler(&mockMCPServersRepo{}, nil)

	body := []byte(`{"decision": "maybe", "approver": "admin"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers/server_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Approve(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMCPServersHandler_Approve_ServerNotFound(t *testing.T) {
	t.Parallel()

	handler := NewMCPServersHandler(&mockMCPServersRepo{found: false}, nil)

	body := []byte(`{"decision": "approved", "approver": "admin"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers/nonexistent/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Approve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMCPServersHandler_Approve_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewMCPServersHandler(&mockMCPServersRepo{}, nil)

	body := []byte(`{"decision":}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers/server_001/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Approve(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

type mockMCPServersRepo struct {
	server  MCPServerDTO
	found   bool
	updated bool
}

func (m *mockMCPServersRepo) GetMCPServer(id string) (MCPServerDTO, bool) {
	if m.found == false && m.server.ID == "" {
		return MCPServerDTO{}, false
	}
	return m.server, true
}

func (m *mockMCPServersRepo) UpdateMCPServer(server MCPServerDTO) error {
	m.updated = true
	m.server = server
	return nil
}