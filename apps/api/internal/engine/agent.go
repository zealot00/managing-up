package engine

import (
	"context"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/server"
)

// Agent executes a Task using available Tools, returning a result and full trace.
type Agent interface {
	Run(ctx context.Context, task server.Task, tools []Tool) (*ExecutionResult, error)
}

// Tool is a callable function available to the Agent.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// ExecutionResult is the output of an Agent.Run() call.
type ExecutionResult struct {
	Output     string        `json:"output"`
	Trace      []AgentStep   `json:"trace"`
	TotalSteps int           `json:"total_steps"`
	Cost       CostInfo      `json:"cost"`
	Duration   time.Duration `json:"duration"`
	Status     string        `json:"status"` // "succeeded" | "failed" | "timeout" | "max_steps_exceeded"
}

// CostInfo tracks LLM token usage for an agent run.
type CostInfo struct {
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Model        string `json:"model"`
}

// AgentStep records one turn of the agent loop.
type AgentStep struct {
	StepNumber  int          `json:"step_number"`
	Messages    []Message    `json:"messages"`
	LLMResponse string       `json:"llm_response"`
	ToolCalls   []ToolCall   `json:"tool_calls"`
	ToolResults []ToolResult `json:"tool_results"`
	Error       string       `json:"error,omitempty"`
}

// Message represents a single message in the LLM conversation.
type Message struct {
	Role       string `json:"role"` // "user" | "assistant" | "tool"
	Content    string `json:"content"`
	ToolName   string `json:"tool_name,omitempty"`
	ToolResult any    `json:"tool_result,omitempty"`
}

// ToolCall is a parsed tool invocation from LLM output.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
	ToolCallID string        `json:"tool_call_id"`
	Output     any           `json:"output"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// Status values for ExecutionResult.Status
const (
	StatusSucceeded        = "succeeded"
	StatusFailed           = "failed"
	StatusTimeout          = "timeout"
	StatusMaxStepsExceeded = "max_steps_exceeded"
)
