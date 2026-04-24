package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"
)

func TestMCPInvokeHandler_Invoke(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{
		hasPermission: true,
		findServer:   true,
		server: MCPServerDTO{
			ID:           "server_001",
			Name:         "test_server",
			TransportType: "stdio",
			URL:          "/path/to/server",
			Command:      "node",
			Args:         []string{"server.js"},
			Env:          []string{},
			Headers:      []string{},
		},
	}, &mockMCPInvoker{
		result: &MCPInvokeResult{
			Success: true,
			Output:  map[string]interface{}{"result": "ok"},
		},
	})

	body := []byte(`{
		"server_id": "server_001",
		"tool_name": "test_tool",
		"parameters": {"key": "value"}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{}, &mockMCPInvoker{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp/invoke", nil)
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_InvalidContentType(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{}, &mockMCPInvoker{})

	body := []byte(`{"server_id": "s1", "tool_name": "t1"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_MissingServerID(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{}, &mockMCPInvoker{})

	body := []byte(`{"tool_name": "test_tool"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_MissingToolName(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{}, &mockMCPInvoker{})

	body := []byte(`{"server_id": "server_001"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_PermissionDenied(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{
		hasPermission: false,
	}, &mockMCPInvoker{})

	body := []byte(`{"server_id": "server_001", "tool_name": "test_tool"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_ServerNotFound(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{
		hasPermission: true,
		findServer:   false,
	}, &mockMCPInvoker{})

	body := []byte(`{"server_id": "nonexistent", "tool_name": "test_tool"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMCPInvokeHandler_Invoke_InvokeFailed(t *testing.T) {
	t.Parallel()

	handler := NewMCPInvokeHandler(&mockMCPInvokeRepo{
		hasPermission: true,
		findServer:   true,
		server: MCPServerDTO{
			ID:           "server_001",
			Name:         "test_server",
			TransportType: "stdio",
		},
	}, &mockMCPInvoker{
		invokeErr: errors.New("invoke failed"),
	})

	body := []byte(`{"server_id": "server_001", "tool_name": "test_tool"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Invoke(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGrantMCPHandler_Grant(t *testing.T) {
	t.Parallel()

	repo := &mockMCPGrantRepo{
		findServer: true,
		server: MCPServerDTO{
			ID:   "server_001",
			Name: "test_server",
		},
	}
	handler := NewGrantMCPHandler(repo)

	body := []byte(`{
		"mcp_server_id": "server_001",
		"user_id": "user_001",
		"permission_type": "invoke"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/grant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if !repo.granted {
		t.Fatalf("expected CreateMCPServerPermission to be called")
	}
}

func TestGrantMCPHandler_Grant_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp/grant", nil)
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestGrantMCPHandler_Grant_InvalidContentType(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	body := []byte(`{"mcp_server_id": "s1"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/grant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestGrantMCPHandler_Grant_MissingServerID(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	body := []byte(`{"user_id": "user_001"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/grant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGrantMCPHandler_Grant_MissingTarget(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	body := []byte(`{"mcp_server_id": "server_001"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/grant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGrantMCPHandler_Grant_ServerNotFound(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{
		findServer: false,
	})

	body := []byte(`{"mcp_server_id": "nonexistent", "user_id": "user_001"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/grant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Grant(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGrantMCPHandler_ListPermissions(t *testing.T) {
	t.Parallel()

	repo := &mockMCPGrantRepo{
		permissions: []MCPServerPermission{
			{
				ID:              "perm_001",
				MCPServerID:    "server_001",
				UserID:         "user_001",
				PermissionType: "invoke",
				IsGranted:      true,
				GrantedAt:      time.Now(),
			},
		},
	}
	handler := NewGrantMCPHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp/permissions?mcp_server_id=server_001", nil)
	rec := httptest.NewRecorder()

	handler.ListPermissions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json response: %v", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data object in response")
	}

	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected items array in response")
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(items))
	}
}

func TestGrantMCPHandler_ListPermissions_MissingServerID(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp/permissions", nil)
	rec := httptest.NewRecorder()

	handler.ListPermissions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGrantMCPHandler_ListPermissions_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewGrantMCPHandler(&mockMCPGrantRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/permissions?mcp_server_id=server_001", nil)
	rec := httptest.NewRecorder()

	handler.ListPermissions(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestParseHeadersToMap(t *testing.T) {
	headers := []string{"Content-Type: application/json", "Authorization: Bearer token"}

	result := parseHeadersToMap(headers)

	if result["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type to be application/json, got %s", result["Content-Type"])
	}
	if result["Authorization"] != "Bearer token" {
		t.Fatalf("expected Authorization to be Bearer token, got %s", result["Authorization"])
	}
}

func TestParseHeadersToMap_InvalidFormat(t *testing.T) {
	headers := []string{"InvalidHeader", "Another:one:two"}

	result := parseHeadersToMap(headers)

	if len(result) != 1 {
		t.Fatalf("expected 1 header, got %d", len(result))
	}
}

type mockMCPInvokeRepo struct {
	hasPermission bool
	findServer    bool
	server        MCPServerDTO
}

func (m *mockMCPInvokeRepo) CheckMCPPermission(mcpServerID, userID, apiKeyID, skillID string) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockMCPInvokeRepo) IncrementMCPRouterCatalogUseCount(serverID string) error {
	return nil
}

func (m *mockMCPInvokeRepo) GetMCPServer(id string) (MCPServerDTO, bool) {
	if m.findServer == false {
		return MCPServerDTO{}, false
	}
	return m.server, true
}

type mockMCPInvoker struct {
	result      *MCPInvokeResult
	invokeErr   error
}

func (m *mockMCPInvoker) InvokeTool(ctx context.Context, config MCPServerConfig, toolName string, params map[string]interface{}) (*MCPInvokeResult, error) {
	if m.invokeErr != nil {
		return nil, m.invokeErr
	}
	return m.result, nil
}

type mockMCPGrantRepo struct {
	server      MCPServerDTO
	findServer  bool
	permissions []MCPServerPermission
	granted     bool
}

func (m *mockMCPGrantRepo) CreateMCPServerPermission(p MCPServerPermission) (MCPServerPermission, error) {
	m.granted = true
	p.ID = "new_perm_id"
	return p, nil
}

func (m *mockMCPGrantRepo) ListMCPServerPermissions(mcpServerID string) ([]MCPServerPermission, error) {
	return m.permissions, nil
}

func (m *mockMCPGrantRepo) GetMCPServer(id string) (MCPServerDTO, bool) {
	if m.findServer == false {
		return MCPServerDTO{}, false
	}
	return m.server, true
}