package generator

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

const (
	OnFailureActionMarkFailed = "mark_failed"
	OnFailureActionContinue   = "continue"
)

const (
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusSucceeded = "succeeded"
	StepStatusFailed    = "failed"
)
