package tool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// Registry manages available tools and their definitions.
type Registry struct {
	tools map[string]Tool
	defs  map[string]*ToolDefinition
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
		defs:  make(map[string]*ToolDefinition),
	}
}

// Register registers a tool with its definition.
// For builtin tools, def may be nil.
func (r *Registry) Register(t Tool, def *ToolDefinition) error {
	name := t.Name()
	if name == "" {
		return &Error{Code: "EMPTY_NAME", Message: "tool name cannot be empty"}
	}

	r.tools[name] = t
	if def != nil {
		def.Name = name // ensure consistency
		r.defs[name] = def
	}

	return nil
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// GetDefinition retrieves a tool definition by name.
func (r *Registry) GetDefinition(name string) (*ToolDefinition, bool) {
	def, ok := r.defs[name]
	return def, ok
}

// List returns all registered tool definitions.
func (r *Registry) List() []*ToolDefinition {
	defs := make([]*ToolDefinition, 0, len(r.defs))
	for _, def := range r.defs {
		defs = append(defs, def)
	}
	return defs
}

// ListTools returns all registered tools.
func (r *Registry) ListTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Discover scans a directory for tool definitions and registers them.
// Expects each tool in its own subdirectory with a manifest.yaml file.
func (r *Registry) Discover(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return &Error{Code: "DISCOVER_FAILED", Message: "failed to read directory", Err: err}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolDir := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(toolDir, "manifest.yaml")

		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue // skip directories without manifest
		}

		// Load and register
		def, err := LoadDefinition(manifestPath)
		if err != nil {
			continue // skip invalid manifests
		}

		impl, err := NewToolImplementation(def)
		if err != nil {
			continue // skip unsupported implementations
		}

		r.defs[def.Name] = def
		r.tools[def.Name] = impl
	}

	return nil
}

// LoadFromManifest loads tool definitions from a manifest file.
func (r *Registry) LoadFromManifest(path string) error {
	def, err := LoadDefinition(path)
	if err != nil {
		return err
	}

	impl, err := NewToolImplementation(def)
	if err != nil {
		return err
	}

	r.defs[def.Name] = def
	r.tools[def.Name] = impl
	return nil
}

// Executor executes tools from the registry.
type Executor struct {
	registry *Registry
}

// NewExecutor creates a new tool executor.
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute runs a tool by name with the given arguments.
func (e *Executor) Execute(ctx context.Context, name string, args map[string]any) (*CallResult, error) {
	tool, ok := e.registry.Get(name)
	if !ok {
		return &CallResult{
			Tool:    name,
			Success: false,
			Error:   "tool not found",
		}, &Error{Code: "TOOL_NOT_FOUND", Message: "tool not found: " + name}
	}

	return ExecuteTool(ctx, tool, args)
}

// ExecuteTool runs a tool with the given arguments.
func ExecuteTool(ctx context.Context, t Tool, args map[string]any) (*CallResult, error) {
	result, err := t.Execute(ctx, args)
	if err != nil {
		return &CallResult{
			Tool:    t.Name(),
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &CallResult{
		Tool:    t.Name(),
		Success: true,
		Output:  result,
	}, nil
}

// CLIExecutor executes a CLI command.
type CLIExecutor struct{}

func (e *CLIExecutor) ExecuteCommand(ctx context.Context, binaryPath string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = filepath.Dir(binaryPath)
	return cmd.Output()
}
