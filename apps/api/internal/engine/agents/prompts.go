package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zealot/managing-up/apps/api/internal/engine"
)

// BuildSystemPrompt creates the system prompt that informs the LLM about available tools.
func BuildSystemPrompt(toolsDescription string) string {
	return fmt.Sprintf(`You are a helpful AI agent with access to the following tools:

Available tools:
%s

Rules:
- When you need to perform an action, respond with a JSON block containing tool calls
- Format: {"tool_calls":[{"id":"call_1","name":"tool_name","arguments":{...}}]}
- After receiving tool results, continue reasoning and call more tools if needed
- When you have the final answer, respond with plain text (no tool calls in JSON format)
- Always use the tool call format even for a single call`, toolsDescription)
}

// ParseToolCalls extracts tool calls from LLM output text.
// Looks for JSON blocks containing "tool_calls".
// Returns nil if no tool calls found (meaning output is the final answer).
func ParseToolCalls(text string) []engine.ToolCall {
	// 1. Find JSON block
	start := strings.Index(text, "```json")
	if start == -1 {
		start = strings.Index(text, "```")
	}
	if start == -1 {
		// No code block found — treat entire text as plain response
		return nil
	}
	end := strings.Index(text[start+7:], "```")
	if end == -1 {
		return nil
	}
	jsonStr := strings.TrimSpace(text[start+7 : start+7+end])

	// 2. Parse JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil
	}

	tcAny, ok := parsed["tool_calls"]
	if !ok {
		return nil
	}
	tcList, ok := tcAny.([]any)
	if !ok {
		return nil
	}

	var calls []engine.ToolCall
	for i, c := range tcList {
		cMap, ok := c.(map[string]any)
		if !ok {
			continue
		}
		call := engine.ToolCall{
			ID:        fmt.Sprintf("call_%d", i+1),
			Name:      "",
			Arguments: make(map[string]any),
		}
		if id, ok := cMap["id"].(string); ok {
			call.ID = id
		}
		if name, ok := cMap["name"].(string); ok {
			call.Name = name
		}
		if args, ok := cMap["arguments"].(map[string]any); ok {
			call.Arguments = args
		}
		calls = append(calls, call)
	}
	return calls
}
