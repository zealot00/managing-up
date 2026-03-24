package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type EventType string

const (
	EventExecutionStarted   EventType = "execution_started"
	EventExecutionSucceeded EventType = "execution_succeeded"
	EventExecutionFailed    EventType = "execution_failed"
	EventStepStarted        EventType = "step_started"
	EventStepSucceeded      EventType = "step_succeeded"
	EventStepFailed         EventType = "step_failed"
	EventLLMCall            EventType = "llm_call"
	EventToolInput          EventType = "tool_input"
	EventToolOutput         EventType = "tool_output"
	EventApprovalRequested  EventType = "approval_requested"
	EventApprovalResolved   EventType = "approval_resolved"
	EventStateChange        EventType = "state_change"
)

type TraceEvent struct {
	ID          string          `json:"id"`
	ExecutionID string          `json:"execution_id"`
	StepID      string          `json:"step_id,omitempty"`
	EventType   EventType       `json:"event_type"`
	EventData   json.RawMessage `json:"event_data"`
	Timestamp   time.Time       `json:"timestamp"`
}

type TraceEmitter interface {
	Emit(ctx context.Context, event TraceEvent) error
}

type TraceRepository interface {
	CreateTrace(event TraceEvent) error
}

type DBTraceEmitter struct {
	db TraceRepository
}

type noOpEmitter struct{}

func NewDBTraceEmitter(db TraceRepository) *DBTraceEmitter {
	return &DBTraceEmitter{db: db}
}

func (e *DBTraceEmitter) Emit(ctx context.Context, event TraceEvent) error {
	if e.db == nil {
		return nil
	}
	return e.db.CreateTrace(event)
}

func (e *noOpEmitter) Emit(ctx context.Context, event TraceEvent) error {
	return nil
}

type LogTraceEmitter struct {
	emit func(event TraceEvent)
}

func NewLogTraceEmitter(emit func(event TraceEvent)) *LogTraceEmitter {
	return &LogTraceEmitter{emit: emit}
}

func (e *LogTraceEmitter) Emit(ctx context.Context, event TraceEvent) error {
	e.emit(event)
	return nil
}

func BuildEventData(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func MustBuildEventData(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return data
}

type ExecutionStartedData struct {
	SkillID   string         `json:"skill_id"`
	SkillName string         `json:"skill_name"`
	Input     map[string]any `json:"input"`
	Triggered string         `json:"triggered_by"`
}

type StepEventData struct {
	StepID    string `json:"step_id"`
	StepType  string `json:"step_type"`
	ToolRef   string `json:"tool_ref,omitempty"`
	AttemptNo int    `json:"attempt_no"`
}

type ToolEventData struct {
	StepID     string         `json:"step_id"`
	ToolRef    string         `json:"tool_ref"`
	Input      map[string]any `json:"input,omitempty"`
	Output     map[string]any `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
	DurationMs int64          `json:"duration_ms"`
}

type ApprovalEventData struct {
	StepID         string `json:"step_id"`
	ApproverGroup  string `json:"approver_group"`
	Message        string `json:"message,omitempty"`
	Decision       string `json:"decision,omitempty"`
	ApprovedBy     string `json:"approved_by,omitempty"`
	ResolutionNote string `json:"resolution_note,omitempty"`
}

type StateChangeData struct {
	From string `json:"from"`
	To   string `json:"to"`
	Step string `json:"step,omitempty"`
}

type LLMCallData struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt,omitempty"`
	Messages    []string `json:"messages,omitempty"`
	InputTokens int      `json:"input_tokens,omitempty"`
	Output      string   `json:"output,omitempty"`
	DurationMs  int64    `json:"duration_ms,omitempty"`
}

func GenerateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}
