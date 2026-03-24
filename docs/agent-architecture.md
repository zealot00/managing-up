# Agent Architecture Design

**Status:** Draft
**Owner:** managing-up
**Target:** Phase P1.1

---

## 1. Goal

Separate the **Agent** (multi-turn tool-call loop) from the **Evaluator** (scoring logic). Agents execute tasks and produce traces. Evaluators judge the quality of outputs. This makes both easier to test, extend, and replace independently.

---

## 2. Core Interfaces

### 2.1 Agent

```go
// Agent executes a Task using available Tools, returning a result and trace.
// It manages the multi-turn loop: LLM → parse tool calls → execute tools → repeat.
type Agent interface {
    Run(ctx context.Context, task Task, tools []Tool) (*ExecutionResult, error)
}
```

### 2.2 Tool

```go
// Tool is a callable function available to the Agent.
type Tool interface {
    Name() string
    Description() string
    // Execute runs the tool with the given arguments.
    // Returns a serializable result or an error.
    Execute(ctx context.Context, args map[string]any) (any, error)
}
```

### 2.3 ExecutionResult

```go
// ExecutionResult is the output of an Agent.Run() call.
type ExecutionResult struct {
    Output     string       // Final text output (when no more tool calls)
    Trace      []AgentStep  // Every turn of the loop
    TotalSteps int
    Cost       CostInfo
    Duration   time.Duration
    Status     string // "succeeded" | "failed" | "timeout" | "max_steps_exceeded"
}

// CostInfo tracks LLM token usage.
type CostInfo struct {
    InputTokens  int
    OutputTokens int
    Model       string
}
```

### 2.4 AgentStep

```go
// AgentStep records one turn of the agent loop.
type AgentStep struct {
    StepNumber   int

    // LLM call
    Messages     []Message  // full conversation up to and including this turn's LLM request
    LLMResponse  string     // raw LLM output text

    // Parsed tool calls (zero or more per turn)
    ToolCalls    []ToolCall

    // Execution results
    ToolResults  []ToolResult

    // Errors
    Error        error
}

// Message represents a single message in the LLM conversation.
type Message struct {
    Role    string // "user" | "assistant" | "tool"
    Content string
    // For tool messages
    ToolName string
    ToolResult any
}

// ToolCall is a parsed tool invocation from LLM output.
type ToolCall struct {
    ID       string
    Name     string
    Arguments map[string]any
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
    ToolCallID string
    Output     any
    Error      error
    Duration   time.Duration
}
```

---

## 3. LLMAgent Implementation

`engine/agents/llm_agent.go`

### 3.1 Loop

```
1. Build initial messages from task input
2. maxTurns times:
   a. Call LLM with current messages
   b. Parse LLM output for tool calls
   c. If no tool calls → return output as ExecutionResult.Output
   d. Execute each tool call in parallel (or sequential)
   e. Append tool results as "tool" role messages
   f. Continue
3. If maxTurns exceeded → return last LLM output as result
```

### 3.2 Tool Call Parsing

LLM outputs text. We parse tool calls from it:

```
1. Look for JSON block in output: ```json ... ```
2. Extract name + arguments from JSON
3. Verify tool name exists in registry
```

Fallback: if no tool calls detected, treat LLM output as final text output.

### 3.3 Tool Registry

```go
type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool)
func (r *ToolRegistry) Get(name string) (Tool, bool)
func (r *ToolRegistry) ListAll() []Tool
```

### 3.4 Context Window Management

Each turn appends assistant message + tool results to messages array. After N turns the context may exceed model limit. Mitigation:
- Track cumulative token count (approximate)
- If approaching limit, truncate oldest non-system messages
- Or: fail with `max_context_exceeded`

---

## 4. How EvaluationRunner Uses Agent

Before (T0.2):
```
RunTask(task) → callLLM(task.Input) → score(output)
```

After:
```
RunTask(task):
  1. tools := registry.ListAll()  // get available tools
  2. result, err := agent.Run(ctx, task, tools)
  3. score(result.Output)
  4. emit trace events for each AgentStep
```

`EvaluationRunner` does NOT care how agent produces the output. It only scores the final text.

---

## 5. Built-in Tools (Phase P1.1)

For MVP, implement these tools to demonstrate the loop:

| Tool | Description | Args |
|------|-------------|------|
| `web_search` | Search the web | `{"query": string}` |
| `calculator` | Evaluate a math expression | `{"expression": string}` |
| `code_executor` | Execute Python/JS code | `{"code": string, "language": string}` |

Tool implementations live in `engine/tools/`.

---

## 6. Database Schema Additions

### 6.1 Execution Traces

New table `execution_steps`:

```sql
CREATE TABLE execution_steps (
    id            TEXT PRIMARY KEY,
    execution_id  TEXT NOT NULL REFERENCES executions(id),
    step_number   INTEGER NOT NULL,
    llm_input     JSONB,
    llm_output    TEXT,
    cost_info     JSONB,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
```

Tool calls and results embedded in `llm_input` (as messages array) and `llm_output`.

### 6.2 Tool Executions

```sql
CREATE TABLE tool_executions (
    id            TEXT PRIMARY KEY,
    step_id      TEXT NOT NULL REFERENCES execution_steps(id),
    tool_name    TEXT NOT NULL,
    arguments    JSONB,
    output       JSONB,
    duration_ms  INTEGER,
    error        TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 7. HTTP API Changes

### 7.1 Trigger Experiment Run

```
POST /api/v1/experiments/{id}/run
```

Triggers `ExperimentService.RunExperiment()`. Each task run produces an `ExecutionResult` with full trace.

### 7.2 Get Execution with Trace

```
GET /api/v1/executions/{id}
```

Returns execution with `steps[]` array showing each agent turn.

---

## 8. File Structure

```
apps/api/internal/engine/
  agent.go                    # Interfaces: Agent, Tool, ExecutionResult, AgentStep
  tool.go                    # Base tool helpers
  tool_registry.go           # ToolRegistry
  agents/
    llm_agent.go            # LLMAgent implementation
    prompts.go              # System prompts, tool call parsing
  tools/
    web_search.go           # web_search tool
    calculator.go           # calculator tool
    code_executor.go       # code_executor tool
```

---

## 9. Migration

```
migrations/0008_agent_architecture.up.sql
migrations/0008_agent_architecture.down.sql
```

---

## 10. Acceptance Criteria

- [ ] `Agent` interface defined and exported from `engine` package
- [ ] `LLMAgent.Run()` completes multi-turn loop with at least 2 turns
- [ ] Tool calls are parsed from LLM output and executed
- [ ] Tool results are appended to LLM context and next turn uses them
- [ ] `EvaluationRunner.RunTask()` delegates to agent, scores final output
- [ ] Full `ExecutionResult.Trace` is stored and queryable via API
- [ ] At least one built-in tool (calculator) works end-to-end
- [ ] `POST /experiments/{id}/run` completes and returns real scores
