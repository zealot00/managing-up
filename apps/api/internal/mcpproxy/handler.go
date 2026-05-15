package mcpproxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// MCPRegistry provides access to backend MCP servers.
type MCPRegistry interface {
	GetClientByServerName(name string) (MCPBackendClient, bool)
	ListMCPServers() []string
	ListTools() []MCPTool
}

// MCPBackendClient is a backend MCP server client.
type MCPBackendClient interface {
	CallTool(ctx context.Context, toolName string, args map[string]any) (*MCPToolResult, error)
	ListResources(ctx context.Context) ([]mcp.Resource, error)
	ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error)
	ListPrompts(ctx context.Context) ([]mcp.Prompt, error)
	GetPrompt(ctx context.Context, name string, args map[string]string) (*mcp.GetPromptResult, error)
}

// MCPToolResult is the result from a backend tool call.
type MCPToolResult struct {
	Content           []mcp.Content
	IsError           bool
	StructuredContent any
}

// MCPTool provides tool metadata.
type MCPTool interface {
	Name() string
	Description() string
	ToolName() string
	ServerName() string
	MCPToolDef() mcp.Tool
}

// handleToolCall proxies a tool call to the backend MCP server.
func (p *ProxyServer) handleToolCall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return nil, fmt.Errorf("unauthorized: no valid authentication")
	}

	// Parse namespaced tool name: "serverName:toolName"
	parts := strings.SplitN(req.Params.Name, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tool name format: %s (expected serverName:toolName)", req.Params.Name)
	}
	serverName, toolName := parts[0], parts[1]

	// Resolve server name to server ID for permission check
	serverID, ok := p.serverIndex.GetMCPServerByName(serverName)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	// Permission check
	hasPermission, err := p.permChecker.CheckMCPPermission(serverID, principal.UserID, principal.APIKeyID, "")
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("permission denied for server %s", serverName)
	}

	// Get backend MCP client
	client, ok := p.registry.GetClientByServerName(serverName)
	if !ok {
		return nil, fmt.Errorf("backend server not connected: %s", serverName)
	}

	// Forward the tool call
	args, _ := req.Params.Arguments.(map[string]any)
	result, err := client.CallTool(ctx, toolName, args)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: fmt.Sprintf("Tool call failed: %v", err)},
			},
			IsError: true,
		}, nil
	}

	if result.IsError {
		return &mcp.CallToolResult{
			Content:           convertContent(result.Content),
			IsError:           true,
			StructuredContent: result.StructuredContent,
		}, nil
	}

	return &mcp.CallToolResult{
		Content:           convertContent(result.Content),
		StructuredContent: result.StructuredContent,
	}, nil
}

// handleListResources proxies a list resources request to all permitted backend servers.
func (p *ProxyServer) handleListResources(ctx context.Context, req mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return &mcp.ListResourcesResult{}, nil
	}

	allowedServers := p.getAllowedServers(ctx, principal)
	var allResources []mcp.Resource

	for serverName := range allowedServers {
		client, ok := p.registry.GetClientByServerName(serverName)
		if !ok {
			continue
		}
		resources, err := client.ListResources(ctx)
		if err != nil {
			continue
		}
		for _, r := range resources {
			r.URI = serverName + ":" + r.URI
			allResources = append(allResources, r)
		}
	}

	return &mcp.ListResourcesResult{Resources: allResources}, nil
}

// handleReadResource proxies a read resource request to the backend server.
func (p *ProxyServer) handleReadResource(ctx context.Context, req mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	parts := strings.SplitN(req.Params.URI, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid resource URI format: %s", req.Params.URI)
	}
	serverName, uri := parts[0], parts[1]

	serverID, ok := p.serverIndex.GetMCPServerByName(serverName)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	hasPermission, err := p.permChecker.CheckMCPPermission(serverID, principal.UserID, principal.APIKeyID, "")
	if err != nil || !hasPermission {
		return nil, fmt.Errorf("permission denied")
	}

	client, ok := p.registry.GetClientByServerName(serverName)
	if !ok {
		return nil, fmt.Errorf("backend server not connected: %s", serverName)
	}

	return client.ReadResource(ctx, uri)
}

// handleListPrompts proxies a list prompts request to all permitted backend servers.
func (p *ProxyServer) handleListPrompts(ctx context.Context, req mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return &mcp.ListPromptsResult{}, nil
	}

	allowedServers := p.getAllowedServers(ctx, principal)
	var allPrompts []mcp.Prompt

	for serverName := range allowedServers {
		client, ok := p.registry.GetClientByServerName(serverName)
		if !ok {
			continue
		}
		prompts, err := client.ListPrompts(ctx)
		if err != nil {
			continue
		}
		for _, pr := range prompts {
			pr.Name = serverName + ":" + pr.Name
			allPrompts = append(allPrompts, pr)
		}
	}

	return &mcp.ListPromptsResult{Prompts: allPrompts}, nil
}

// handleGetPrompt proxies a get prompt request to the backend server.
func (p *ProxyServer) handleGetPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	parts := strings.SplitN(req.Params.Name, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid prompt name format: %s", req.Params.Name)
	}
	serverName, promptName := parts[0], parts[1]

	serverID, ok := p.serverIndex.GetMCPServerByName(serverName)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	hasPermission, err := p.permChecker.CheckMCPPermission(serverID, principal.UserID, principal.APIKeyID, "")
	if err != nil || !hasPermission {
		return nil, fmt.Errorf("permission denied")
	}

	client, ok := p.registry.GetClientByServerName(serverName)
	if !ok {
		return nil, fmt.Errorf("backend server not connected: %s", serverName)
	}

	return client.GetPrompt(ctx, promptName, req.Params.Arguments)
}

// getAllowedServers returns the set of server names the principal has permission to use.
func (p *ProxyServer) getAllowedServers(ctx context.Context, principal *Principal) map[string]bool {
	perms, err := p.permChecker.ListPermissionsForIdentity(principal.UserID, principal.APIKeyID)
	if err != nil {
		return nil
	}

	allowed := make(map[string]bool)
	for _, perm := range perms {
		for _, sn := range p.registry.ListMCPServers() {
			if id, ok := p.serverIndex.GetMCPServerByName(sn); ok && id == perm.MCPServerID {
				allowed[sn] = true
			}
		}
	}
	return allowed
}

// convertContent converts []mcp.Content from the backend client result.
func convertContent(contents []mcp.Content) []mcp.Content {
	if contents == nil {
		return nil
	}
	result := make([]mcp.Content, len(contents))
	for i, c := range contents {
		result[i] = c
	}
	return result
}
