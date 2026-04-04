package executors

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	tool "github.com/zealot/managing-up/apps/api/internal/engine/tool"
)

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
