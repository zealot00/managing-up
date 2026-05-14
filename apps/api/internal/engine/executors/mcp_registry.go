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

// toolNamespacedKey returns the namespaced registry key for a tool: "serverName:toolName".
func toolNamespacedKey(serverName, toolName string) string {
	return serverName + ":" + toolName
}

type MCPRegistry struct {
	clients *MCPClients
	mu      sync.RWMutex
	tools   map[string]tool.Tool // key = "serverName:toolName" (namespaced)
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

	// Wire disconnect callback
	if config.OnDisconnect != nil {
		serverName := name
		client.OnConnectionLost(func(err error) {
			config.OnDisconnect(serverName)
		})
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

	// Remove all tools belonging to this server
	for key, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok && mt.serverName == name {
			delete(r.tools, key)
		}
	}

	return nil
}

func (r *MCPRegistry) registerTools(serverName string, client *MCPClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// First, remove any existing tools for this server (handles re-registration)
	for key, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok && mt.serverName == serverName {
			delete(r.tools, key)
		}
	}

	for _, t := range client.ListTools() {
		mcpTool := NewMCPTool(serverName, t, client)
		key := toolNamespacedKey(serverName, mcpTool.ToolName())
		r.tools[key] = mcpTool
	}
}

// GetToolByServer looks up a tool by server name and tool name (namespace-aware).
func (r *MCPRegistry) GetToolByServer(serverName, toolName string) (tool.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := toolNamespacedKey(serverName, toolName)
	t, ok := r.tools[key]
	return t, ok
}

// GetTool looks up a tool by namespaced key "serverName:toolName".
// For backward compatibility, if the key contains no ":", it searches across all servers.
func (r *MCPRegistry) GetTool(name string) (tool.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If namespaced key, look up directly
	if strings.Contains(name, ":") {
		t, ok := r.tools[name]
		return t, ok
	}

	// Fallback: search across all servers by bare tool name (backward compat)
	for _, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok && mt.ToolName() == name {
			return t, true
		}
	}
	return nil, false
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

	seen := make(map[string]bool)
	servers := make([]string, 0)
	for _, t := range r.tools {
		if mt, ok := t.(*MCPTool); ok && !seen[mt.serverName] {
			seen[mt.serverName] = true
			servers = append(servers, mt.serverName)
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

// IsRegistered returns whether a client with the given name is already registered.
func (r *MCPRegistry) IsRegistered(name string) bool {
	_, ok := r.clients.GetClient(name)
	return ok
}

// GetClientByServerName returns the MCPClient for the given server name.
func (r *MCPRegistry) GetClientByServerName(name string) (*MCPClient, bool) {
	return r.clients.GetClient(name)
}

// HealthCheck pings all registered MCP servers and returns name→error map.
func (r *MCPRegistry) HealthCheck(ctx context.Context) map[string]error {
	servers := r.ListMCPServers()
	results := make(map[string]error, len(servers))
	for _, name := range servers {
		c, ok := r.clients.GetClient(name)
		if !ok {
			results[name] = fmt.Errorf("client not found")
			continue
		}
		results[name] = c.Ping(ctx)
	}
	return results
}

// RefreshTools re-fetches the tool list for a server and updates the registry.
func (r *MCPRegistry) RefreshTools(ctx context.Context, serverName string) error {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return fmt.Errorf("MCP client not found: %s", serverName)
	}

	// Re-list tools from the server
	listResult, err := c.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	// Update client's tool cache
	c.mu.Lock()
	newTools := make(map[string]mcp.Tool, len(listResult.Tools))
	for _, t := range listResult.Tools {
		newTools[t.Name] = t
	}
	c.tools = newTools
	c.mu.Unlock()

	// Re-register tools in the registry
	r.registerTools(serverName, c)
	return nil
}

// ListResourcesByServer returns resources for a specific server.
func (r *MCPRegistry) ListResourcesByServer(ctx context.Context, serverName string) ([]mcp.Resource, error) {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.ListResources(ctx)
}

// ReadResource reads a resource from a specific server.
func (r *MCPRegistry) ReadResource(ctx context.Context, serverName, uri string) (*mcp.ReadResourceResult, error) {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.ReadResource(ctx, uri)
}

// ListResourceTemplatesByServer returns resource templates for a specific server.
func (r *MCPRegistry) ListResourceTemplatesByServer(ctx context.Context, serverName string) ([]mcp.ResourceTemplate, error) {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.ListResourceTemplates(ctx)
}

// SubscribeResource subscribes to resource changes on a specific server.
func (r *MCPRegistry) SubscribeResource(ctx context.Context, serverName, uri string) error {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.Subscribe(ctx, uri)
}

// UnsubscribeResource unsubscribes from resource changes on a specific server.
func (r *MCPRegistry) UnsubscribeResource(ctx context.Context, serverName, uri string) error {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.Unsubscribe(ctx, uri)
}

// ListPromptsByServer returns prompts for a specific server.
func (r *MCPRegistry) ListPromptsByServer(ctx context.Context, serverName string) ([]mcp.Prompt, error) {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.ListPrompts(ctx)
}

// GetPrompt retrieves a prompt from a specific server.
func (r *MCPRegistry) GetPrompt(ctx context.Context, serverName, name string, args map[string]string) (*mcp.GetPromptResult, error) {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("MCP client not found: %s", serverName)
	}
	return c.GetPrompt(ctx, name, args)
}

// GetServerCapabilities returns the capabilities of a specific server.
func (r *MCPRegistry) GetServerCapabilities(serverName string) *mcp.ServerCapabilities {
	c, ok := r.clients.GetClient(serverName)
	if !ok {
		return nil
	}
	return c.GetServerCapabilities()
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
