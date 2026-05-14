package executors

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Command allowlist for stdio transport - restrict to known safe commands
var allowedCommands = map[string]bool{
	"npx":     true,
	"node":    true,
	"python":  true,
	"python3": true,
	"uv":      true,
	"docker":  true,
	"kubectl": true,
	"gh":      true,
	"git":     true,
	"curl":    true,
	"wget":    true,
}

// headerRegex validates HTTP headers don't contain CRLF injection
var headerRegex = regexp.MustCompile(`^[\w-]+:[\w\s\-./:;,]+$`)

// MCPClientConfig configures an MCP client
type MCPClientConfig struct {
	Command      string
	Args         []string
	Env          []string
	Timeout      time.Duration
	AllowedPaths []string // Additional allowed command paths
	OnDisconnect func(serverName string)
}

// validateStdioConfig validates stdio transport configuration for security
func validateStdioConfig(config MCPClientConfig) error {
	if config.Command == "" {
		return fmt.Errorf("command is required for stdio transport")
	}

	cmd := config.Command
	parts := strings.Split(cmd, " ")
	base := parts[0]
	baseName := strings.TrimPrefix(base, "/usr/local/bin/")
	baseName = strings.TrimPrefix(baseName, "/usr/bin/")
	baseName = strings.Split(baseName, "/")[0]

	if !allowedCommands[baseName] {
		if strings.HasPrefix(cmd, "/") {
			allowed := false
			for _, allowedPath := range config.AllowedPaths {
				if strings.HasPrefix(cmd, allowedPath) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("command not in allowlist: %s", cmd)
			}
		} else {
			return fmt.Errorf("command not in allowlist: %s", cmd)
		}
	}

	shellMeta := regexp.MustCompile("[;&|`\"$<>\\\\]")
	for _, arg := range config.Args {
		if shellMeta.MatchString(arg) {
			return fmt.Errorf("arg contains shell metacharacters: %s", arg)
		}
	}

	return nil
}

// validateHTTPHeaders validates headers don't contain CRLF injection
func validateHTTPHeaders(headers map[string]string) error {
	for key, value := range headers {
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("header contains invalid characters: %s: %s", key, value)
		}
		if !headerRegex.MatchString(key + ":" + value) {
			return fmt.Errorf("header format invalid: %s: %s", key, value)
		}
	}
	return nil
}

type MCPClient struct {
	name   string
	config MCPClientConfig
	client *client.Client
	tools  map[string]mcp.Tool
	mu     sync.Mutex // Mutex for state changes
	closed bool
}

type MCPClients struct {
	clients map[string]*MCPClient
	mu      sync.RWMutex
}

func NewMCPClients() *MCPClients {
	return &MCPClients{
		clients: make(map[string]*MCPClient),
	}
}

func (m *MCPClients) Register(ctx context.Context, name string, config MCPClientConfig) error {
	// Phase 1: Validate BEFORE acquiring lock
	if err := validateStdioConfig(config); err != nil {
		return fmt.Errorf("invalid stdio config: %w", err)
	}

	// Phase 2: Quick existence check WITHOUT write lock
	m.mu.RLock()
	if _, exists := m.clients[name]; exists {
		m.mu.RUnlock()
		return fmt.Errorf("MCP client already registered: %s", name)
	}
	m.mu.RUnlock()

	// Phase 3: Network I/O WITHOUT holding lock (prevents fat lock)
	var initCtx context.Context
	var initCancel context.CancelFunc
	if config.Timeout > 0 {
		initCtx, initCancel = context.WithTimeout(ctx, config.Timeout)
		defer initCancel()
	} else {
		initCtx = ctx
	}

	var t transport.Interface
	t = transport.NewStdio(config.Command, config.Args, config.Env...)

	mcpClient := client.NewClient(t)
	if err := mcpClient.Start(initCtx); err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Ensure cleanup on error after Start succeeds
	cleanup := func() {
		if mcpClient != nil {
			mcpClient.Close()
		}
	}

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
		cleanup()
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Verify protocol version
	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		cleanup()
		return fmt.Errorf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION)
	}

	listTools, err := mcpClient.ListTools(initCtx, mcp.ListToolsRequest{})
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make(map[string]mcp.Tool)
	for _, tool := range listTools.Tools {
		tools[tool.Name] = tool
	}

	// Phase 4: Acquire lock ONLY for map insertion (nanoseconds)
	m.mu.Lock()
	if _, exists := m.clients[name]; exists {
		m.mu.Unlock()
		cleanup()
		return fmt.Errorf("MCP client already registered: %s", name)
	}
	m.clients[name] = &MCPClient{
		name:   name,
		config: config,
		client: mcpClient,
		tools:  tools,
	}
	m.mu.Unlock()

	return nil
}

func (m *MCPClients) RegisterHTTP(ctx context.Context, name string, baseURL string, headers map[string]string) error {
	// Phase 1: Validate headers BEFORE any operations
	if err := validateHTTPHeaders(headers); err != nil {
		return fmt.Errorf("invalid HTTP headers: %w", err)
	}

	// Phase 2: Quick existence check WITHOUT write lock
	m.mu.RLock()
	if _, exists := m.clients[name]; exists {
		m.mu.RUnlock()
		return fmt.Errorf("MCP client already registered: %s", name)
	}
	m.mu.RUnlock()

	// Phase 3: Network I/O WITHOUT holding lock (prevents fat lock)
	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	defer initCancel()

	t, err := transport.NewStreamableHTTP(baseURL, transport.WithHTTPHeaders(headers))
	if err != nil {
		return fmt.Errorf("failed to create HTTP transport: %w", err)
	}

	mcpClient := client.NewClient(t)
	if err := mcpClient.Start(initCtx); err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Ensure cleanup on error after Start succeeds
	cleanup := func() {
		if mcpClient != nil {
			mcpClient.Close()
		}
	}

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
		cleanup()
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Verify protocol version
	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		cleanup()
		return fmt.Errorf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION)
	}

	listTools, err := mcpClient.ListTools(initCtx, mcp.ListToolsRequest{})
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make(map[string]mcp.Tool)
	for _, tool := range listTools.Tools {
		tools[tool.Name] = tool
	}

	// Phase 4: Acquire lock ONLY for map insertion (nanoseconds)
	m.mu.Lock()
	if _, exists := m.clients[name]; exists {
		m.mu.Unlock()
		cleanup()
		return fmt.Errorf("MCP client already registered: %s", name)
	}
	m.clients[name] = &MCPClient{
		name:   name,
		client: mcpClient,
		tools:  tools,
	}
	m.mu.Unlock()

	return nil
}

func (m *MCPClients) GetClient(name string) (*MCPClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[name]
	return c, ok
}

func (m *MCPClients) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.clients[name]
	if !ok {
		return fmt.Errorf("MCP client not found: %s", name)
	}

	// Close the client
	if c.client != nil {
		c.client.Close()
	}

	delete(m.clients, name)
	return nil
}

func (m *MCPClients) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, c := range m.clients {
		if c.client != nil {
			c.client.Close()
		}
		delete(m.clients, name)
	}
}

func (c *MCPClient) Name() string {
	return c.name
}

func (c *MCPClient) ListTools() []mcp.Tool {
	c.mu.Lock()
	defer c.mu.Unlock()

	tools := make([]mcp.Tool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (c *MCPClient) GetTool(name string) (mcp.Tool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	tool, ok := c.tools[name]
	return tool, ok
}

// CallTool executes a tool with proper synchronization.
// Lock is held for the entire duration of the call to prevent use-after-close.
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("client closed")
	}

	if _, ok := c.tools[name]; !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Check if context is already cancelled before making the call
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
}

// Close closes the MCP client with proper synchronization
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping checks if the MCP server is reachable.
func (c *MCPClient) Ping(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	return c.client.Ping(ctx)
}

// GetServerCapabilities returns the server capabilities discovered during initialization.
func (c *MCPClient) GetServerCapabilities() *mcp.ServerCapabilities {
	capabilities := c.client.GetServerCapabilities()
	return &capabilities
}

// ListResources returns the resources available on the MCP server.
func (c *MCPClient) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	result, err := c.client.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return nil, err
	}
	return result.Resources, nil
}

// ReadResource reads a specific resource from the MCP server.
func (c *MCPClient) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	return c.client.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	})
}

// ListResourceTemplates returns the resource templates available on the MCP server.
func (c *MCPClient) ListResourceTemplates(ctx context.Context) ([]mcp.ResourceTemplate, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	result, err := c.client.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
	if err != nil {
		return nil, err
	}
	return result.ResourceTemplates, nil
}

// Subscribe subscribes to resource changes.
func (c *MCPClient) Subscribe(ctx context.Context, uri string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	return c.client.Subscribe(ctx, mcp.SubscribeRequest{
		Params: mcp.SubscribeParams{
			URI: uri,
		},
	})
}

// Unsubscribe unsubscribes from resource changes.
func (c *MCPClient) Unsubscribe(ctx context.Context, uri string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	return c.client.Unsubscribe(ctx, mcp.UnsubscribeRequest{
		Params: mcp.UnsubscribeParams{
			URI: uri,
		},
	})
}

// ListPrompts returns the prompts available on the MCP server.
func (c *MCPClient) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	result, err := c.client.ListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		return nil, err
	}
	return result.Prompts, nil
}

// GetPrompt retrieves a specific prompt from the MCP server.
func (c *MCPClient) GetPrompt(ctx context.Context, name string, args map[string]string) (*mcp.GetPromptResult, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	c.mu.Unlock()

	return c.client.GetPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      name,
			Arguments: args,
		},
	})
}

// OnNotification registers a handler for server-initiated notifications.
func (c *MCPClient) OnNotification(handler func(mcp.JSONRPCNotification)) {
	c.client.OnNotification(handler)
}

// OnConnectionLost registers a handler for connection loss events.
func (c *MCPClient) OnConnectionLost(handler func(error)) {
	c.client.OnConnectionLost(handler)
}

// execLookPath validates a command exists in PATH
func execLookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("command not found in PATH: %s", name)
	}
	return path, nil
}
