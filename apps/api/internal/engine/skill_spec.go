package engine

type SkillSpec struct {
	Name        string     `yaml:"name"`
	Version     string     `yaml:"version"`
	RiskLevel   string     `yaml:"risk_level"`
	Description string     `yaml:"description"`
	Inputs      []Input    `yaml:"inputs"`
	Steps       []Step     `yaml:"steps"`
	OnFailure   *OnFailure `yaml:"on_failure"`
}

type Input struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
}

type Step struct {
	ID             string            `yaml:"id"`
	Type           string            `yaml:"type"`
	ToolRef        string            `yaml:"tool_ref,omitempty"`
	With           map[string]string `yaml:"with,omitempty"`
	TimeoutSeconds int               `yaml:"timeout_seconds,omitempty"`
	ApproverGroup  string            `yaml:"approver_group,omitempty"`
	Message        string            `yaml:"message,omitempty"`
	RetryPolicy    *RetryPolicy      `yaml:"retry_policy,omitempty"`
	Condition      string            `yaml:"condition,omitempty"`
}

type ApprovalStep struct {
	ID            string `yaml:"id"`
	Type          string `yaml:"type"`
	ApproverGroup string `yaml:"approver_group"`
	Message       string `yaml:"message"`
}

type ToolStep struct {
	ID             string            `yaml:"id"`
	Type           string            `yaml:"type"`
	ToolRef        string            `yaml:"tool_ref"`
	With           map[string]string `yaml:"with"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	RetryPolicy    *RetryPolicy      `yaml:"retry_policy,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts    int `yaml:"max_attempts"`
	BackoffSeconds int `yaml:"backoff_seconds"`
}

type OnFailure struct {
	Action string `yaml:"action"`
}

type ExecutionStep struct {
	ID          string         `json:"id"`
	ExecutionID string         `json:"execution_id"`
	StepID      string         `json:"step_id"`
	Status      string         `json:"status"`
	ToolRef     string         `json:"tool_ref,omitempty"`
	StartedAt   string         `json:"started_at"`
	EndedAt     string         `json:"ended_at,omitempty"`
	DurationMs  int64          `json:"duration_ms,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	AttemptNo   int            `json:"attempt_no"`
}

const (
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusSucceeded = "succeeded"
	StepStatusFailed    = "failed"
)

const (
	OnFailureActionMarkFailed = "mark_failed"
	OnFailureActionContinue   = "continue"
)
