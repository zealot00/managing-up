package engine

import (
	"fmt"
	"strings"
)

// ToolRegistry manages available tools for an agent.
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]Tool)}
}

func (r *ToolRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *ToolRegistry) ListAll() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

func (r *ToolRegistry) BuildToolsDescription() string {
	var parts []string
	for _, t := range r.tools {
		parts = append(parts, fmt.Sprintf("- %s: %s", t.Name(), t.Description()))
	}
	return strings.Join(parts, "\n")
}
