package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine/executors"
)

type MCPRegistryAdapter struct {
	registry *executors.MCPRegistry
}

func NewMCPRegistryAdapter(registry *executors.MCPRegistry) *MCPRegistryAdapter {
	return &MCPRegistryAdapter{registry: registry}
}

func (s *MCPRegistryAdapter) InvokeTool(ctx context.Context, config MCPServerConfig, toolName string, params map[string]interface{}) (*MCPInvokeResult, error) {
	if s.registry == nil {
		return &MCPInvokeResult{Success: false, Error: "MCP registry not initialized"}, nil
	}

	var args map[string]any = params
	if args == nil {
		args = make(map[string]any)
	}

	// Register the client if not already registered
	if !s.registry.IsRegistered(config.Name) {
		switch config.TransportType {
		case "http", "https":
			if err := s.registry.RegisterHTTP(ctx, config.Name, config.URL, config.Headers); err != nil {
				return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("failed to register HTTP client: %v", err)}, nil
			}
		case "stdio":
			env := config.Env
			if env == nil {
				env = []string{}
			}
			if err := s.registry.RegisterStdio(ctx, config.Name, executors.MCPClientConfig{
				Command: config.Command,
				Args:    config.Args,
				Env:     env,
				Timeout: 30 * time.Second,
			}); err != nil {
				return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("failed to register stdio client: %v", err)}, nil
			}
		default:
			return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("unsupported transport type: %s", config.TransportType)}, nil
		}
	}

	result, err := s.invokeToolByName(ctx, config.Name, toolName, args)
	if err != nil {
		return &MCPInvokeResult{Success: false, Error: err.Error()}, nil
	}

	return &MCPInvokeResult{
		Success: true,
		Output:  result,
	}, nil
}

func (s *MCPRegistryAdapter) invokeToolByName(ctx context.Context, serverName, toolName string, args map[string]any) (map[string]any, error) {
	// Use namespace-aware lookup: server + tool name
	t, ok := s.registry.GetToolByServer(serverName, toolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s on server %s", toolName, serverName)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	if m, ok := result.(map[string]any); ok {
		return m, nil
	}
	return map[string]any{"result": result}, nil
}
