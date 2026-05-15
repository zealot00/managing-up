package mcpproxy

import (
	"context"
	"log"
	"net/http"
	"strings"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
)

// ServerIndex resolves server names to server IDs.
type ServerIndex interface {
	GetMCPServerByName(name string) (id string, ok bool)
}

// ProxyServer is an MCP protocol proxy that sits between MCP clients
// and backend MCP servers. It handles authentication, permission filtering,
// and request forwarding.
type ProxyServer struct {
	mcpServer   *mcpserver.MCPServer
	httpServer  *mcpserver.StreamableHTTPServer
	registry    MCPRegistry
	keyResolver APIKeyResolver
	permChecker PermChecker
	serverIndex ServerIndex
}

// NewProxyServer creates a new MCP proxy server.
func NewProxyServer(
	registry MCPRegistry,
	keyResolver APIKeyResolver,
	permChecker PermChecker,
	serverIndex ServerIndex,
) *ProxyServer {
	p := &ProxyServer{
		registry:    registry,
		keyResolver: keyResolver,
		permChecker: permChecker,
		serverIndex: serverIndex,
	}

	// Create MCP server with tool filter
	p.mcpServer = mcpserver.NewMCPServer(
		"skill-hub-mcp-proxy",
		"1.0.0",
		mcpserver.WithToolFilter(p.toolFilter),
	)

	// Register all tools from the registry
	p.registerTools()

	// Create StreamableHTTP server with auth context injection
	p.httpServer = mcpserver.NewStreamableHTTPServer(
		p.mcpServer,
		mcpserver.WithHTTPContextFunc(p.authContextFunc),
	)

	return p
}

// Handler returns the http.Handler for the MCP proxy endpoint.
func (p *ProxyServer) Handler() http.Handler {
	return p.httpServer
}

// RefreshTools re-registers all tools from the registry.
// Call this after a new MCP server is approved at runtime.
func (p *ProxyServer) RefreshTools() {
	p.registerTools()
}

// authContextFunc extracts the Bearer token from the HTTP request,
// resolves it to a Principal, and stores it in the context.
func (p *ProxyServer) authContextFunc(ctx context.Context, r *http.Request) context.Context {
	authHeader := r.Header.Get("Authorization")
	rawKey := extractBearerToken(authHeader)

	// Also check x-api-key header
	if rawKey == "" {
		rawKey = r.Header.Get("x-api-key")
	}

	if rawKey == "" || p.keyResolver == nil {
		return ctx
	}

	dbKeyID, userID, ok := p.keyResolver.ResolveAPIKey(rawKey)
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, principalKey, &Principal{
		APIKeyID: dbKeyID,
		UserID:   userID,
	})
}

// registerTools registers all tools from the MCP registry as proxy tools.
func (p *ProxyServer) registerTools() {
	if p.registry == nil {
		return
	}

	tools := p.registry.ListTools()
	var serverTools []mcpserver.ServerTool

	for _, t := range tools {
		proxyTool := mcp.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
			},
		}

		toolDef := t.MCPToolDef()
		if len(toolDef.InputSchema.Properties) > 0 {
			proxyTool.InputSchema.Properties = toolDef.InputSchema.Properties
		}
		if len(toolDef.InputSchema.Required) > 0 {
			proxyTool.InputSchema.Required = toolDef.InputSchema.Required
		}
		if toolDef.InputSchema.Type != "" {
			proxyTool.InputSchema.Type = toolDef.InputSchema.Type
		}

		serverTools = append(serverTools, mcpserver.ServerTool{
			Tool:    proxyTool,
			Handler: p.handleToolCall,
		})
	}

	if len(serverTools) > 0 {
		p.mcpServer.AddTools(serverTools...)
	}

	log.Printf("MCP proxy: registered %d tools", len(serverTools))

	p.registerResources()
	p.registerPrompts()
}

func (p *ProxyServer) registerResources() {
	p.mcpServer.AddResource(mcp.Resource{
		Name:        "mcp-resources",
		Description: "Resources from all connected MCP servers",
		URI:         "mcp://resources",
	}, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		result, err := p.handleListResources(ctx, mcp.ListResourcesRequest{})
		if err != nil {
			return nil, err
		}
		contents := make([]mcp.ResourceContents, len(result.Resources))
		for i, r := range result.Resources {
			contents[i] = mcp.TextResourceContents{
				URI:      r.URI,
				MIMEType: "text/plain",
				Text:     r.Name,
			}
		}
		return contents, nil
	})
}

func (p *ProxyServer) registerPrompts() {
	p.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "mcp-prompts",
		Description: "Prompts from all connected MCP servers",
	}, p.handleGetPrompt)
}

// registryAdapter wraps *executors.MCPRegistry to implement MCPRegistry interface.
type registryAdapter struct {
	inner *executors.MCPRegistry
}

func NewRegistryAdapter(inner *executors.MCPRegistry) MCPRegistry {
	return &registryAdapter{inner: inner}
}

func (a *registryAdapter) GetClientByServerName(name string) (MCPBackendClient, bool) {
	client, ok := a.inner.GetClientByServerName(name)
	if !ok {
		return nil, false
	}
	return &backendClientAdapter{client: client}, true
}

func (a *registryAdapter) ListMCPServers() []string {
	return a.inner.ListMCPServers()
}

func (a *registryAdapter) ListTools() []MCPTool {
	tools := a.inner.ListTools()
	result := make([]MCPTool, len(tools))
	for i, t := range tools {
		if mt, ok := t.(*executors.MCPTool); ok {
			result[i] = &mcpToolAdapter{tool: mt}
		}
	}
	return result
}

type backendClientAdapter struct {
	client *executors.MCPClient
}

func (a *backendClientAdapter) CallTool(ctx context.Context, toolName string, args map[string]any) (*MCPToolResult, error) {
	result, err := a.client.CallTool(ctx, toolName, args)
	if err != nil {
		return nil, err
	}
	return &MCPToolResult{
		Content:           result.Content,
		IsError:           result.IsError,
		StructuredContent: result.StructuredContent,
	}, nil
}

func (a *backendClientAdapter) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return a.client.ListResources(ctx)
}

func (a *backendClientAdapter) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	return a.client.ReadResource(ctx, uri)
}

func (a *backendClientAdapter) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	return a.client.ListPrompts(ctx)
}

func (a *backendClientAdapter) GetPrompt(ctx context.Context, name string, args map[string]string) (*mcp.GetPromptResult, error) {
	return a.client.GetPrompt(ctx, name, args)
}

type mcpToolAdapter struct {
	tool *executors.MCPTool
}

func (a *mcpToolAdapter) Name() string         { return a.tool.Name() }
func (a *mcpToolAdapter) Description() string   { return a.tool.Description() }
func (a *mcpToolAdapter) ToolName() string      { return a.tool.ToolName() }
func (a *mcpToolAdapter) ServerName() string     { return a.tool.ServerName() }
func (a *mcpToolAdapter) MCPToolDef() mcp.Tool  { return a.tool.MCPToolDef() }

// Ensure raw key extraction handles various formats
func init() {
	_ = strings.TrimSpace
}
