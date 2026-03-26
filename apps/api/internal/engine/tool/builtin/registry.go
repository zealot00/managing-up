package builtin

import (
	"github.com/zealot/managing-up/apps/api/internal/engine/tool"
)

// Registry holds all builtin tools.
type Registry struct {
	tools map[string]tool.Tool
}

// NewRegistry creates a new builtin tool registry.
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]tool.Tool)}
	r.registerDefaultTools()
	return r
}

func (r *Registry) registerDefaultTools() {
	r.Register(NewCalculator())
}

// Get retrieves a builtin tool by name.
func (r *Registry) Get(name string) (tool.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all builtin tools.
func (r *Registry) List() []tool.Tool {
	tools := make([]tool.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Register adds a tool to the registry.
func (r *Registry) Register(t tool.Tool) {
	r.tools[t.Name()] = t
}
