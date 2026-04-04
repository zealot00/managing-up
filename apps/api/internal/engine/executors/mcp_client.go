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

	// Check allowlist
	base := strings.TrimPrefix(config.Command, "/usr/local/bin/")
	base = strings.TrimPrefix(base, "/usr/bin/")
	base = strings.Split(base, " ")[0] // Handle full paths

	if !allowedCommands[base] && !strings.HasPrefix(config.Command, "/") {
		return fmt.Errorf("command not in allowlist: %s", config.Command)
	}

	// Validate args don't contain shell metacharacters
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
	// Validate configuration BEFORE any operations
	if err := validateStdioConfig(config); err != nil {
		return fmt.Errorf("invalid stdio config: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[name]; exists {
		return fmt.Errorf("MCP client already registered: %s", name)
	}

	// Apply timeout if configured
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	var t transport.Interface
	t = transport.NewStdio(config.Command, config.Args, config.Env...)

	mcpClient := client.NewClient(t)
	if err := mcpClient.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Ensure cleanup on error after Start succeeds
	cleanup := func() {
		if mcpClient != nil {
			mcpClient.Close()
		}
	}

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
		cleanup()
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Verify protocol version
	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		cleanup()
		return fmt.Errorf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION)
	}

	listTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make(map[string]mcp.Tool)
	for _, tool := range listTools.Tools {
		tools[tool.Name] = tool
	}

	m.clients[name] = &MCPClient{
		name:   name,
		config: config,
		client: mcpClient,
		tools:  tools,
	}

	return nil
}

func (m *MCPClients) RegisterHTTP(ctx context.Context, name string, baseURL string, headers map[string]string) error {
	// Validate headers BEFORE any operations
	if err := validateHTTPHeaders(headers); err != nil {
		return fmt.Errorf("invalid HTTP headers: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[name]; exists {
		return fmt.Errorf("MCP client already registered: %s", name)
	}

	// Apply timeout if configured
	if defaultTimeout := 30 * time.Second; true {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	t, err := transport.NewStreamableHTTP(baseURL, transport.WithHTTPHeaders(headers))
	if err != nil {
		return fmt.Errorf("failed to create HTTP transport: %w", err)
	}

	mcpClient := client.NewClient(t)
	if err := mcpClient.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Ensure cleanup on error after Start succeeds
	cleanup := func() {
		if mcpClient != nil {
			mcpClient.Close()
		}
	}

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
		cleanup()
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Verify protocol version
	if initialized.ProtocolVersion != mcp.LATEST_PROTOCOL_VERSION {
		cleanup()
		return fmt.Errorf("protocol version mismatch: got %s, want %s", initialized.ProtocolVersion, mcp.LATEST_PROTOCOL_VERSION)
	}

	listTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make(map[string]mcp.Tool)
	for _, tool := range listTools.Tools {
		tools[tool.Name] = tool
	}

	m.clients[name] = &MCPClient{
		name:   name,
		client: mcpClient,
		tools:  tools,
	}

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

// execLookPath validates a command exists in PATH
func execLookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("command not found in PATH: %s", name)
	}
	return path, nil
}
