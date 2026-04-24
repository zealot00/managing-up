package handlers

import (
	"context"
	"fmt"

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

	var result map[string]any
	var err error

	switch config.TransportType {
	case "http", "https":
		err = s.registry.RegisterHTTP(ctx, config.Name, config.URL, config.Headers)
		if err != nil {
			return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("failed to register HTTP client: %v", err)}, nil
		}
		result, err = s.invokeToolByName(ctx, config.Name, toolName, args)
	case "stdio":
		env := config.Env
		if env == nil {
			env = []string{}
		}
		err = s.registry.RegisterStdio(ctx, config.Name, executors.MCPClientConfig{
			Command: config.Command,
			Args:    config.Args,
			Env:     env,
			Timeout: 30,
		})
		if err != nil {
			return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("failed to register stdio client: %v", err)}, nil
		}
		result, err = s.invokeToolByName(ctx, config.Name, toolName, args)
	default:
		return &MCPInvokeResult{Success: false, Error: fmt.Sprintf("unsupported transport type: %s", config.TransportType)}, nil
	}

	if err != nil {
		return &MCPInvokeResult{Success: false, Error: err.Error()}, nil
	}

	return &MCPInvokeResult{
		Success: true,
		Output:  result,
	}, nil
}

func (s *MCPRegistryAdapter) invokeToolByName(ctx context.Context, serverName, toolName string, args map[string]any) (map[string]any, error) {
	t, ok := s.registry.GetTool(toolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
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