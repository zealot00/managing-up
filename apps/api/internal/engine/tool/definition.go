package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Error represents a tool-related error.
type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// LoadDefinition loads a tool definition from a YAML manifest file.
func LoadDefinition(path string) (*ToolDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &Error{Code: "LOAD_FAILED", Message: "failed to read file", Err: err}
	}

	var def ToolDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, &Error{Code: "PARSE_FAILED", Message: "failed to parse manifest", Err: err}
	}

	if def.Name == "" {
		return nil, &Error{Code: "INVALID_DEFINITION", Message: "tool name is required"}
	}

	return &def, nil
}

// NewToolImplementation creates an executable tool from a definition.
func NewToolImplementation(def *ToolDefinition) (Tool, error) {
	switch def.Impl.Type {
	case "cli":
		return NewCLITool(def), nil
	case "http":
		return NewHTTPAdapter(def), nil
	case "script":
		return NewScriptTool(def), nil
	case "builtin":
		return nil, &Error{Code: "UNSUPPORTED", Message: "builtin tools must be registered separately"}
	default:
		return nil, &Error{Code: "UNKNOWN_TYPE", Message: "unknown tool implementation type: " + def.Impl.Type}
	}
}

// ToolCategory categorizes a tool type.
type ToolCategory string

const (
	CategoryCLI     ToolCategory = "cli"
	CategoryHTTP    ToolCategory = "http"
	CategoryScript  ToolCategory = "script"
	CategoryBuiltin ToolCategory = "builtin"
)

// RiskLevel indicates the risk of executing a tool.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// ToolDefinition describes a tool registered in the system.
// Builtin tools (calculator, echo, etc.) do not use this.
type ToolDefinition struct {
	Name     string       `json:"name"`
	Category ToolCategory `json:"category"`
	Version  string       `json:"version"` // current version for rollback comparison

	Description string     `json:"description"`
	InputSchema JSONSchema `json:"input_schema,omitempty"`

	// Security
	RequiresApproval bool      `json:"requires_approval"`
	RiskLevel        RiskLevel `json:"risk_level,omitempty"`

	// Implementation
	Impl ToolImplementation `json:"impl,omitempty"`
}

// ToolImplementation describes how a tool is actually executed.
type ToolImplementation struct {
	// Type: "cli" | "http" | "script" | "builtin"
	Type string `json:"type"`

	// CLI
	BinaryPath string `json:"binary_path,omitempty"`
	Command    string `json:"command,omitempty"`

	// HTTP
	BaseURL string `json:"base_url,omitempty"`

	// Script
	Interpreter string `json:"interpreter,omitempty"`
	ScriptPath  string `json:"script_path,omitempty"`

	// Builtin
	BuiltinName string `json:"builtin_name,omitempty"`
}

// JSONSchema describes input/output schema for a tool.
type JSONSchema struct {
	Type       string                `json:"type,omitempty"`
	Properties map[string]SchemaProp `json:"properties,omitempty"`
	Required   []string              `json:"required,omitempty"`
}

// SchemaProp describes a single property in a JSON schema.
type SchemaProp struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`
	Example     any    `json:"example,omitempty"`
}

// ToolCall represents a single tool invocation in execution trace.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// CallResult is the result of executing a tool.
type CallResult struct {
	Tool     string         `json:"tool"`
	Action   string         `json:"action,omitempty"`
	Success  bool           `json:"success"`
	Output   any            `json:"output,omitempty"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
