package mcpproxy

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// PermEntry represents a granted permission.
type PermEntry struct {
	MCPServerID string
}

// PermChecker checks MCP permissions.
type PermChecker interface {
	CheckMCPPermission(serverID, userID, apiKeyID, skillID string) (bool, error)
	ListPermissionsForIdentity(userID, apiKeyID string) ([]PermEntry, error)
}

// toolFilter filters tools based on the authenticated principal's permissions.
// This is used as a ToolFilterFunc for the MCP server.
func (p *ProxyServer) toolFilter(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		// No auth = no tools visible
		return nil
	}

	// Get allowed server set
	allowedServers := p.getAllowedServers(ctx, principal)
	if len(allowedServers) == 0 {
		return nil
	}

	// Filter tools: only include tools from allowed servers
	var filtered []mcp.Tool
	for _, tool := range tools {
		parts := strings.SplitN(tool.Name, ":", 2)
		if len(parts) == 2 {
			serverName := parts[0]
			if allowedServers[serverName] {
				filtered = append(filtered, tool)
			}
		}
	}
	return filtered
}
