package mcpproxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolFilter_NoAuth(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{}
	tools := []mcp.Tool{
		{Name: "github:search", Description: "Search"},
	}

	filtered := p.toolFilter(context.Background(), tools)
	if filtered != nil {
		t.Fatalf("expected nil for unauthenticated request, got %d tools", len(filtered))
	}
}

func TestToolFilter_NoPermissions(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		permChecker: &mockPermChecker{perms: []PermEntry{}},
		serverIndex: &mockServerIndex{servers: map[string]string{}},
		registry:    &mockRegistry{servers: []string{}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	tools := []mcp.Tool{
		{Name: "github:search", Description: "Search"},
	}

	filtered := p.toolFilter(ctx, tools)
	if filtered != nil {
		t.Fatalf("expected nil when no permissions, got %d tools", len(filtered))
	}
}

func TestToolFilter_WithPermissions(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		permChecker: &mockPermChecker{perms: []PermEntry{{MCPServerID: "server_001"}}},
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001", "jira": "server_002"}},
		registry:    &mockRegistry{servers: []string{"github", "jira"}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	tools := []mcp.Tool{
		{Name: "github:search", Description: "Search GitHub"},
		{Name: "jira:search", Description: "Search Jira"},
		{Name: "github:create", Description: "Create Issue"},
	}

	filtered := p.toolFilter(ctx, tools)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 tools (github only), got %d", len(filtered))
	}
	for _, tool := range filtered {
		if tool.Name != "github:search" && tool.Name != "github:create" {
			t.Errorf("unexpected tool in filtered list: %s", tool.Name)
		}
	}
}

func TestToolFilter_ToolWithoutNamespace(t *testing.T) {
	t.Parallel()

	p := &ProxyServer{
		permChecker: &mockPermChecker{perms: []PermEntry{{MCPServerID: "server_001"}}},
		serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001"}},
		registry:    &mockRegistry{servers: []string{"github"}},
	}
	ctx := context.WithValue(context.Background(), principalKey, &Principal{APIKeyID: "key1", UserID: "user1"})
	tools := []mcp.Tool{
		{Name: "bare-name", Description: "No namespace"},
	}

	filtered := p.toolFilter(ctx, tools)
	if len(filtered) != 0 {
		t.Fatalf("expected 0 tools for non-namespaced tool, got %d", len(filtered))
	}
}

func TestGetAllowedServers(t *testing.T) {
	t.Parallel()

	t.Run("with matching permissions", func(t *testing.T) {
		p := &ProxyServer{
			permChecker: &mockPermChecker{perms: []PermEntry{{MCPServerID: "server_001"}}},
			serverIndex: &mockServerIndex{servers: map[string]string{"github": "server_001", "jira": "server_002"}},
			registry:    &mockRegistry{servers: []string{"github", "jira"}},
		}
		principal := &Principal{APIKeyID: "key1", UserID: "user1"}

		allowed := p.getAllowedServers(context.Background(), principal)
		if len(allowed) != 1 {
			t.Fatalf("expected 1 allowed server, got %d", len(allowed))
		}
		if !allowed["github"] {
			t.Error("expected github to be allowed")
		}
		if allowed["jira"] {
			t.Error("expected jira to NOT be allowed")
		}
	})

	t.Run("with error", func(t *testing.T) {
		p := &ProxyServer{
			permChecker: &mockPermChecker{permListErr: fmt.Errorf("db error")},
		}
		principal := &Principal{APIKeyID: "key1", UserID: "user1"}

		allowed := p.getAllowedServers(context.Background(), principal)
		if allowed != nil {
			t.Fatal("expected nil on error")
		}
	})
}
