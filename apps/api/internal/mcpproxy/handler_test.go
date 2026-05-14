package mcpproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// --- Mocks ---

type mockAPIKeyResolver struct {
	dbKeyID string
	userID  string
	ok      bool
}

func (m *mockAPIKeyResolver) ResolveAPIKey(rawKey string) (string, string, bool) {
	return m.dbKeyID, m.userID, m.ok
}

type mockPermChecker struct {
	hasPermission    bool
	permErr          error
	perms            []PermEntry
	permListErr      error
}

func (m *mockPermChecker) CheckMCPPermission(serverID, userID, apiKeyID, skillID string) (bool, error) {
	return m.hasPermission, m.permErr
}

func (m *mockPermChecker) ListPermissionsForIdentity(userID, apiKeyID string) ([]PermEntry, error) {
	return m.perms, m.permListErr
}

type mockServerIndex struct {
	servers map[string]string // name -> id
}

func (m *mockServerIndex) GetMCPServerByName(name string) (string, bool) {
	id, ok := m.servers[name]
	return id, ok
}

type mockBackendClient struct {
	callToolResult *MCPToolResult
	callToolErr    error
	resources      []mcp.Resource
	prompts        []mcp.Prompt
}

func (m *mockBackendClient) CallTool(ctx context.Context, toolName string, args map[string]any) (*MCPToolResult, error) {
	return m.callToolResult, m.callToolErr
}

func (m *mockBackendClient) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return m.resources, nil
}

func (m *mockBackendClient) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{}, nil
}

func (m *mockBackendClient) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	return m.prompts, nil
}

func (m *mockBackendClient) GetPrompt(ctx context.Context, name string, args map[string]string) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{}, nil
}

type mockRegistry struct {
	clients map[string]*mockBackendClient
	servers []string
	tools   []MCPTool
}

func (m *mockRegistry) GetClientByServerName(name string) (MCPBackendClient, bool) {
	c, ok := m.clients[name]
	if !ok {
		return nil, false
	}
	return c, true
}

func (m *mockRegistry) ListMCPServers() []string {
	return m.servers
}

func (m *mockRegistry) ListTools() []MCPTool {
	return m.tools
}

type mockMCPTool struct {
	name        string
	description string
	toolName    string
	serverName  string
	toolDef     mcp.Tool
}

func (m *mockMCPTool) Name() string        { return m.name }
func (m *mockMCPTool) Description() string  { return m.description }
func (m *mockMCPTool) ToolName() string     { return m.toolName }
func (m *mockMCPTool) ServerName() string    { return m.serverName }
func (m *mockMCPTool) MCPToolDef() mcp.Tool { return m.toolDef }

// --- Tests ---

func TestHandleToolCall_Unauthorized(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.Background()
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "server:tool"}}

	_, err := p.handleToolCall(ctx, req)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
}

func TestHandleToolCall_InvalidToolName(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "no-colon"}}

	_, err := p.handleToolCall(ctx, req)
	if err == nil {
		t.Fatal("expected error for tool name without colon")
	}
}

func TestHandleToolCall_ServerNotFound(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "unknown-server:tool"}}

	_, err := p.handleToolCall(ctx, req)
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
}

func TestHandleToolCall_PermissionDenied(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		permChecker: &mockPermChecker{hasPermission: false},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "github:search"}}

	_, err := p.handleToolCall(ctx, req)
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}

func TestHandleToolCall_BackendNotConnected(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		permChecker: &mockPermChecker{hasPermission: true},
		registry:    &mockRegistry{clients: map[string]*mockBackendClient{}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "github:search"}}

	_, err := p.handleToolCall(ctx, req)
	if err == nil {
		t.Fatal("expected backend not connected error")
	}
}

func TestHandleToolCall_Success(t *testing.T) {
	t.Parallel()

	client := &mockBackendClient{
		callToolResult: &MCPToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "result"},
			},
		},
	}
	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		permChecker: &mockPermChecker{hasPermission: true},
		registry:    &mockRegistry{clients: map[string]*mockBackendClient{"github": client}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "github:search", Arguments: map[string]any{"q": "test"}}}

	result, err := p.handleToolCall(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected successful result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestHandleToolCall_BackendError(t *testing.T) {
	t.Parallel()

	client := &mockBackendClient{callToolErr: fmt.Errorf("connection refused")}
	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		permChecker: &mockPermChecker{hasPermission: true},
		registry:    &mockRegistry{clients: map[string]*mockBackendClient{"github": client}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "github:search"}}

	result, err := p.handleToolCall(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result from backend failure")
	}
}

func TestHandleToolCall_BackendToolError(t *testing.T) {
	t.Parallel()

	client := &mockBackendClient{
		callToolResult: &MCPToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "something went wrong"},
			},
			IsError: true,
		},
	}
	p := &ProxyServer{
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		permChecker: &mockPermChecker{hasPermission: true},
		registry:    &mockRegistry{clients: map[string]*mockBackendClient{"github": client}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "github:search"}}

	result, err := p.handleToolCall(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected isError=true from backend tool error")
	}
}

func TestHandleReadResource_Unauthorized(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.Background()
	req := mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: "github:file://test"}}

	_, err := p.handleReadResource(ctx, req)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
}

func TestHandleReadResource_InvalidURI(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1"})
	req := mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: "no-colon"}}

	_, err := p.handleReadResource(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid URI")
	}
}

func TestHandleGetPrompt_Unauthorized(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.Background()
	req := mcp.GetPromptRequest{Params: mcp.GetPromptParams{Name: "server:prompt"}}

	_, err := p.handleGetPrompt(ctx, req)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
}

func TestHandleGetPrompt_InvalidName(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1"})
	req := mcp.GetPromptRequest{Params: mcp.GetPromptParams{Name: "no-colon"}}

	_, err := p.handleGetPrompt(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid prompt name")
	}
}

func TestHandleListResources_NoAuth(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.Background()

	result, err := p.handleListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Resources) != 0 {
		t.Fatal("expected empty resources for unauthenticated request")
	}
}

func TestHandleListPrompts_NoAuth(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	ctx := context.Background()

	result, err := p.handleListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Prompts) != 0 {
		t.Fatal("expected empty prompts for unauthenticated request")
	}
}

func TestConvertContent_Nil(t *testing.T) {
	t.Parallel()

	result := convertContent(nil)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestConvertContent_WithContent(t *testing.T) {
	t.Parallel()

	input := []mcp.Content{
		mcp.TextContent{Type: "text", Text: "hello"},
	}
	result := convertContent(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
}

// --- Auth Context Tests ---

func TestAuthContextFunc_BearerToken(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		keyResolver: &mockAPIKeyResolver{dbKeyID: "db_key_1", userID: "user_1", ok: true},
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer skhub_test123")

	ctx := p.authContextFunc(context.Background(), req)

	principal := PrincipalFromContext(ctx)
	if principal == nil {
		t.Fatal("expected principal from Bearer token")
	}
	if principal.APIKeyID != "db_key_1" {
		t.Errorf("expected APIKeyID db_key_1, got %s", principal.APIKeyID)
	}
	if principal.UserID != "user_1" {
		t.Errorf("expected UserID user_1, got %s", principal.UserID)
	}
}

func TestAuthContextFunc_XAPIKey(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		keyResolver: &mockAPIKeyResolver{dbKeyID: "db_key_2", userID: "user_2", ok: true},
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("x-api-key", "skhub_xapi")

	ctx := p.authContextFunc(context.Background(), req)

	principal := PrincipalFromContext(ctx)
	if principal == nil {
		t.Fatal("expected principal from x-api-key header")
	}
	if principal.APIKeyID != "db_key_2" {
		t.Errorf("expected APIKeyID db_key_2, got %s", principal.APIKeyID)
	}
}

func TestAuthContextFunc_NoAuth(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		keyResolver: &mockAPIKeyResolver{},
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)

	ctx := p.authContextFunc(context.Background(), req)

	principal := PrincipalFromContext(ctx)
	if principal != nil {
		t.Fatal("expected nil principal when no auth provided")
	}
}

func TestAuthContextFunc_InvalidKey(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		keyResolver: &mockAPIKeyResolver{ok: false},
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer invalid_key")

	ctx := p.authContextFunc(context.Background(), req)

	principal := PrincipalFromContext(ctx)
	if principal != nil {
		t.Fatal("expected nil principal when key resolution fails")
	}
}

func TestAuthContextFunc_NilResolver(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{keyResolver: nil}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer skhub_test")

	ctx := p.authContextFunc(context.Background(), req)

	principal := PrincipalFromContext(ctx)
	if principal != nil {
		t.Fatal("expected nil principal when keyResolver is nil")
	}
}
