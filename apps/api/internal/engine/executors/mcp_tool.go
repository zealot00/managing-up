package executors

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type MCPTool struct {
	serverName string
	tool       mcp.Tool
	client     *MCPClient
}

func NewMCPTool(serverName string, tool mcp.Tool, client *MCPClient) *MCPTool {
	return &MCPTool{
		serverName: serverName,
		tool:       tool,
		client:     client,
	}
}

// Name returns the namespaced tool name in the format "serverName:toolName".
// This prevents collisions when multiple servers expose tools with the same name.
func (t *MCPTool) Name() string {
	return toolNamespacedKey(t.serverName, t.tool.Name)
}

// ToolName returns the bare tool name (without server prefix).
func (t *MCPTool) ToolName() string {
	return t.tool.Name
}

func (t *MCPTool) Description() string {
	return t.tool.Description
}

// ServerName returns the name of the MCP server this tool belongs to.
func (t *MCPTool) ServerName() string {
	return t.serverName
}

// MCPToolDef returns the underlying mcp.Tool definition.
func (t *MCPTool) MCPToolDef() mcp.Tool {
	return t.tool
}

func (t *MCPTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	result, err := t.client.CallTool(ctx, t.tool.Name, args)
	if err != nil {
		return nil, fmt.Errorf("MCP tool call failed: %w", err)
	}

	if result.IsError {
		return nil, fmt.Errorf("MCP tool error: %s", extractTextContent(result.Content))
	}

	return map[string]any{
		"content":           result.Content,
		"structructuredContent": result.StructuredContent,
	}, nil
}

func extractTextContent(contents []mcp.Content) string {
	var text string
	for _, c := range contents {
		if tc, ok := c.(mcp.TextContent); ok {
			text += tc.Text
		}
	}
	return text
}

type MCPToolInfo struct {
	ServerName  string
	Name        string
	Description string
	InputSchema map[string]any
}

func MCPToolInfoFromMCP(serverName string, tool mcp.Tool) MCPToolInfo {
	schema := map[string]any{}
	if len(tool.InputSchema.Properties) > 0 {
		schema["properties"] = tool.InputSchema.Properties
	}
	if len(tool.InputSchema.Required) > 0 {
		schema["required"] = tool.InputSchema.Required
	}

	return MCPToolInfo{
		ServerName:  serverName,
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: schema,
	}
}
