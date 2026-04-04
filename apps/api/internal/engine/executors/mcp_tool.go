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

func (t *MCPTool) Name() string {
	return t.tool.Name
}

func (t *MCPTool) Description() string {
	return t.tool.Description
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
		"structuredContent": result.StructuredContent,
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
