package executors

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	tool "github.com/zealot/managing-up/apps/api/internal/engine/tool"
)

// MCPServerConfig is the configuration for an MCP server (used for validation).
type MCPServerConfig struct {
	TransportType string
	Command       string
	Args          []string
	Env           []string
	URL           string
	Headers       []string
}

// MCPServerValidationResult represents the result of validating an MCP server.
type MCPServerValidationResult struct {
	Valid bool
	Tools []mcp.Tool
	Error string
}

var globalMCPRegistry *MCPRegistry

type MCPRegistry struct {
	clients *MCPClients
	mu      sync.RWMutex
	tools   map[string]tool.Tool
}

func NewMCPRegistry() *MCPRegistry {
	return &MCPRegistry{
		clients: NewMCPClients(),
		tools:   make(map[string]tool.Tool),
	}
}

func SetGlobalRegistry(r *MCPRegistry) {
	globalMCPRegistry = r
}

func GetGlobalRegistry() *MCPRegistry {
	return globalMCPRegistry
}

func (r *MCPRegistry) RegisterStdio(ctx context.Context, name string, config MCPClientConfig) error {
	if err := r.clients.Register(ctx, name, config); err != nil {
		return fmt.Errorf("failed to register MCP client %s: %w", name, err)
	}

	client, ok := r.clients.GetClient(name)
	if !ok {
		return fmt.Errorf("MCP client %s not found after registration", name)
	}

	r.registerTools(name, client)
	return nil
}

func (r *MCPRegistry) RegisterHTTP(ctx context.Context, name string, baseURL string, headers map[string]string) error {
	if err := r.clients.RegisterHTTP(ctx, name, baseURL, headers); err != nil {
		return fmt.Errorf("failed to register MCP HTTP client %s: %w", name, err)
	}

	client, ok := r.clients.GetClient(name)
	if !ok {
		return fmt.Errorf("MCP client %s not found after registration", name)
	}

	r.registerTools(name, client)
	return nil
}

func (r *MCPRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients.GetClient(name); !ok {
		return fmt.Errorf("MCP client not found: %s", name)
	}

	if err := r.clients.Unregister(name); err != nil {
		return err
	}

	for toolName, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok && mt.serverName == name {
			delete(r.tools, toolName)
		}
	}

	return nil
}

func (r *MCPRegistry) registerTools(serverName string, client *MCPClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, t := range client.ListTools() {
		mcpTool := NewMCPTool(serverName, t, client)
		r.tools[mcpTool.Name()] = mcpTool
	}
}

func (r *MCPRegistry) GetTool(name string) (tool.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

func (r *MCPRegistry) ListTools() []tool.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]tool.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

func (r *MCPRegistry) ListMCPServers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := make([]string, 0)
	for _, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok {
			found := false
			for _, s := range servers {
				if s == mt.serverName {
					found = true
					break
				}
			}
			if !found {
				servers = append(servers, mt.serverName)
			}
		}
	}
	return servers
}

func (r *MCPRegistry) ListToolsByServer(serverName string) []mcp.Tool {
	r.mu.RLock()
	client, ok := r.clients.GetClient(serverName)
	if !ok {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	return client.ListTools()
}

func (r *MCPRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]tool.Tool)
	r.clients.Close()
	return nil
}

func ValidateMCPServer(ctx context.Context, srv MCPServerConfig) MCPServerValidationResult {
	if srv.TransportType == "stdio" {
		return validateStdioServer(ctx, srv)
	}
	if srv.TransportType == "http" || srv.TransportType == "https" || srv.TransportType == "sse" {
		return validateHTTPServer(ctx, srv)
	}
	return MCPServerValidationResult{
		Valid: false,
		Error: fmt.Sprintf("unsupported transport type: %s", srv.TransportType),
	}
}

func validateStdioServer(ctx context.Context, srv MCPServerConfig) MCPServerValidationResult {
	config := MCPClientConfig{
		Command: srv.Command,
		Args:    srv.Args,
		Env:     srv.Env,
		Timeout: 30 * time.Second,
	}

	if err := validateStdioConfig(config); err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("invalid stdio config: %s", err.Error()),
		}
	}

	initCtx, initCancel := context.WithTimeout(ctx, config.Timeout)
	defer initCancel()

	t := transport.NewStdio(config.Command, config.Args, config.Env...)
	mcpClient := client.NewClient(t)

	if err := mcpClient.Start(initCtx); err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to start MCP client: %s", err.Error()),
		}
	}

	cleanup := func() {
		mcpClient.Close()
	}
	defer cleanup()

	initialized, err := mcpClient.Initialize(initCtx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "enterprise-gateway",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to initialize MCP client: %s", err.Error()),
		}
	}

	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION),
		}
	}

	listTools, err := mcpClient.ListTools(initCtx, mcp.ListToolsRequest{})
	if err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to list tools: %s", err.Error()),
		}
	}

	return MCPServerValidationResult{
		Valid: true,
		Tools: listTools.Tools,
	}
}

func validateHTTPServer(ctx context.Context, srv MCPServerConfig) MCPServerValidationResult {
	headers := make(map[string]string)
	for _, h := range srv.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if err := validateHTTPHeaders(headers); err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("invalid HTTP headers: %s", err.Error()),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	t, err := transport.NewStreamableHTTP(srv.URL, transport.WithHTTPHeaders(headers))
	if err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to create HTTP transport: %s", err.Error()),
		}
	}

	mcpClient := client.NewClient(t)

	if err := mcpClient.Start(ctx); err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to start MCP client: %s", err.Error()),
		}
	}

	cleanup := func() {
		mcpClient.Close()
	}
	defer cleanup()

	initialized, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "enterprise-gateway",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to initialize MCP client: %s", err.Error()),
		}
	}

	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION),
		}
	}

	listTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return MCPServerValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to list tools: %s", err.Error()),
		}
	}

	return MCPServerValidationResult{
		Valid: true,
		Tools: listTools.Tools,
	}
}
