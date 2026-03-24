package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/engine"
	"github.com/zealot/managing-up/apps/api/internal/llm"
	"github.com/zealot/managing-up/apps/api/internal/server"
)

const defaultMaxTurns = 10

// LLMAgent implements the engine.Agent interface using an LLM client.
type LLMAgent struct {
	client   llm.Client
	registry *engine.ToolRegistry
	maxTurns int
}

func NewLLMAgent(client llm.Client, registry *engine.ToolRegistry) *LLMAgent {
	return &LLMAgent{
		client:   client,
		registry: registry,
		maxTurns: defaultMaxTurns,
	}
}

// Run executes the agent loop for a given task.
func (a *LLMAgent) Run(ctx context.Context, task server.Task, tools []engine.Tool) (*engine.ExecutionResult, error) {
	start := time.Now()

	// Build tools description and system prompt
	toolsDesc := a.buildToolsDescription(tools)
	systemPrompt := BuildSystemPrompt(toolsDesc)

	// Build initial messages
	messages := a.buildMessages(task, systemPrompt)

	// Convert to llm messages
	llmMessages := a.toLLMMessages(messages)

	var totalCost engine.CostInfo
	totalCost.Model = string(a.client.Model())

	var result engine.ExecutionResult
	result.Trace = []engine.AgentStep{}

	// Multi-turn loop
	for turn := 0; turn < a.maxTurns; turn++ {
		// Call LLM
		resp, err := a.client.Generate(ctx, llmMessages)
		if err != nil {
			result.Trace = append(result.Trace, engine.AgentStep{
				StepNumber:  turn + 1,
				Messages:    messages,
				LLMResponse: "",
				Error:       err.Error(),
			})
			result.Status = engine.StatusFailed
			result.Duration = time.Since(start)
			return &result, err
		}

		// Track cost
		totalCost.InputTokens += int(resp.Usage.InputTokens)
		totalCost.OutputTokens += int(resp.Usage.OutputTokens)

		llmResponseText := resp.Content

		// Parse tool calls from response
		toolCalls := ParseToolCalls(llmResponseText)

		step := engine.AgentStep{
			StepNumber:  turn + 1,
			Messages:    messages,
			LLMResponse: llmResponseText,
			ToolCalls:   toolCalls,
			ToolResults: []engine.ToolResult{},
		}

		// Append assistant message
		messages = append(messages, engine.Message{
			Role:    "assistant",
			Content: llmResponseText,
		})

		if len(toolCalls) == 0 {
			// No tool calls — this is the final answer
			result.Output = llmResponseText
			result.Trace = append(result.Trace, step)
			result.Status = engine.StatusSucceeded
			result.TotalSteps = turn + 1
			result.Cost = totalCost
			result.Duration = time.Since(start)
			return &result, nil
		}

		// Execute tool calls and collect results
		for _, tc := range toolCalls {
			tcStart := time.Now()
			tool, found := a.registry.Get(tc.Name)
			if !found {
				step.ToolResults = append(step.ToolResults, engine.ToolResult{
					ToolCallID: tc.ID,
					Output:     nil,
					Error:      fmt.Sprintf("tool not found: %s", tc.Name),
					Duration:   time.Since(tcStart),
				})
				continue
			}
			execResult, execErr := tool.Execute(ctx, tc.Arguments)
			tr := engine.ToolResult{
				ToolCallID: tc.ID,
				Output:     execResult,
				Duration:   time.Since(tcStart),
			}
			if execErr != nil {
				tr.Error = execErr.Error()
			}
			step.ToolResults = append(step.ToolResults, tr)

			// Append tool result message to conversation
			resultJSON, _ := json.Marshal(execResult)
			messages = append(messages, engine.Message{
				Role:       "tool",
				Content:    string(resultJSON),
				ToolName:   tc.Name,
				ToolResult: execResult,
			})
		}

		result.Trace = append(result.Trace, step)
	}

	// Max turns exceeded
	result.Status = engine.StatusMaxStepsExceeded
	result.TotalSteps = a.maxTurns
	result.Cost = totalCost
	result.Duration = time.Since(start)
	return &result, nil
}

func (a *LLMAgent) buildToolsDescription(tools []engine.Tool) string {
	var parts []string
	for _, t := range tools {
		parts = append(parts, fmt.Sprintf("- %s: %s", t.Name(), t.Description()))
	}
	return strings.Join(parts, "\n")
}

func (a *LLMAgent) buildMessages(task server.Task, systemPrompt string) []engine.Message {
	inputContent := ""

	// Build input from task
	switch task.Input.Source {
	case "inline":
		if len(task.TestCases) > 0 {
			for k, v := range task.TestCases[0].Input {
				inputContent += fmt.Sprintf("%s: %v\n", k, v)
			}
		}
	default:
		inputContent = fmt.Sprintf("Task: %s\nDescription: %s", task.Name, task.Description)
	}
	if inputContent == "" {
		inputContent = task.Description
	}

	return []engine.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: inputContent},
	}
}

func (a *LLMAgent) toLLMMessages(msgs []engine.Message) []llm.Message {
	out := make([]llm.Message, len(msgs))
	for i, m := range msgs {
		out[i] = llm.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return out
}
