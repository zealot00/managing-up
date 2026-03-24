package server

import "time"

// Envelope is the standard API response wrapper.
type Envelope struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
	Meta  Meta      `json:"meta"`
}

// APIError is the normalized API error shape.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta contains response metadata.
type Meta struct {
	RequestID  string      `json:"request_id"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination metadata for list responses.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// Skill represents registry metadata exposed by the API.
type Skill struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OwnerTeam      string    `json:"owner_team"`
	RiskLevel      string    `json:"risk_level"`
	Status         string    `json:"status"`
	CurrentVersion string    `json:"current_version"`
	CreatedBy      string    `json:"created_by,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// SkillVersion represents an immutable version entry for a skill.
type SkillVersion struct {
	ID               string    `json:"id"`
	SkillID          string    `json:"skill_id"`
	Version          string    `json:"version"`
	Status           string    `json:"status"`
	ChangeSummary    string    `json:"change_summary"`
	ApprovalRequired bool      `json:"approval_required"`
	CreatedAt        time.Time `json:"created_at"`
	SpecYaml         string    `json:"spec_yaml,omitempty"`
}

// ProcedureDraft represents a validated or draft SOP extraction result.
type ProcedureDraft struct {
	ID               string    `json:"id"`
	ProcedureKey     string    `json:"procedure_key"`
	Title            string    `json:"title"`
	ValidationStatus string    `json:"validation_status"`
	RequiredTools    []string  `json:"required_tools"`
	SourceType       string    `json:"source_type"`
	CreatedAt        time.Time `json:"created_at"`
}

// Approval represents an approval checkpoint waiting for or recording action.
type Approval struct {
	ID             string    `json:"id"`
	ExecutionID    string    `json:"execution_id"`
	SkillName      string    `json:"skill_name"`
	StepID         string    `json:"step_id"`
	Status         string    `json:"status"`
	ApproverGroup  string    `json:"approver_group"`
	RequestedAt    time.Time `json:"requested_at"`
	ApprovedBy     string    `json:"approved_by,omitempty"`
	ResolutionNote string    `json:"resolution_note,omitempty"`
}

// Execution represents a summarized execution run.
type Execution struct {
	ID            string         `json:"id"`
	SkillID       string         `json:"skill_id"`
	SkillName     string         `json:"skill_name"`
	Status        string         `json:"status"`
	TriggeredBy   string         `json:"triggered_by"`
	StartedAt     time.Time      `json:"started_at"`
	CurrentStepID string         `json:"current_step_id"`
	Input         map[string]any `json:"input,omitempty"`
	CreatedBy     string         `json:"created_by,omitempty"`
}

// DashboardSummary aggregates headline metrics for the control plane.
type DashboardSummary struct {
	ActiveSkills       int     `json:"active_skills"`
	PublishedVersions  int     `json:"published_versions"`
	RunningExecutions  int     `json:"running_executions"`
	WaitingApprovals   int     `json:"waiting_approvals"`
	SuccessRate        float64 `json:"success_rate"`
	AvgDurationSeconds int     `json:"avg_duration_seconds"`
}

// DashboardData is the homepage payload.
type DashboardData struct {
	Summary          DashboardSummary `json:"summary"`
	RecentExecutions []Execution      `json:"recent_executions"`
}

// CreateSkillRequest is the payload for creating a skill.
type CreateSkillRequest struct {
	Name      string `json:"name"`
	OwnerTeam string `json:"owner_team"`
	RiskLevel string `json:"risk_level"`
}

// CreateExecutionRequest is the payload for creating an execution.
type CreateExecutionRequest struct {
	SkillID     string         `json:"skill_id"`
	TriggeredBy string         `json:"triggered_by"`
	Input       map[string]any `json:"input"`
}

// ApproveExecutionRequest is the payload for resolving an approval.
type ApproveExecutionRequest struct {
	Approver string `json:"approver"`
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

// AgentRegistration is the payload for registering an agent.
type AgentRegistration struct {
	AgentID      string   `json:"agent_id"`
	Name         string   `json:"name"`
	Version      string   `json:"version,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	CallbackURL  string   `json:"callback_url,omitempty"`
}

// Agent represents a registered agent.
type Agent struct {
	AgentID      string    `json:"agent_id"`
	Name         string    `json:"name"`
	Version      string    `json:"version,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
	RegisteredAt time.Time `json:"registered_at"`
}

// TraceEvent represents an execution trace record.
type TraceEvent struct {
	ID          string         `json:"id"`
	ExecutionID string         `json:"execution_id"`
	StepID      string         `json:"step_id,omitempty"`
	EventType   string         `json:"event_type"`
	EventData   map[string]any `json:"event_data"`
	Timestamp   time.Time      `json:"timestamp"`
}

// Task represents a reusable evaluation task.
type Task struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`

	// === New fields (Phase 0.1) ===
	TaskType  string          `json:"task_type"` // "benchmark" | "regression" | "ablation"
	Input     TaskInput       `json:"input"`
	Gold      GoldConfig      `json:"gold"`
	Scoring   ScoringConfig   `json:"scoring"`
	Execution ExecutionConfig `json:"execution"`

	// === Legacy fields (kept for backward compatibility) ===
	SkillID    string     `json:"skill_id,omitempty"`
	Difficulty string     `json:"difficulty,omitempty"` // deprecated: use Scoring.Threshold instead
	TestCases  []TestCase `json:"test_cases,omitempty"` // used when Input.Source="inline"
}

// TestCase represents a test case for task evaluation.
type TestCase struct {
	Input    map[string]any `json:"input"`
	Expected any            `json:"expected"`
}

// TaskInput defines the input source for a task.
type TaskInput struct {
	Source string `json:"source"`           // "inline" | "file" | "api" | "synthetic"
	Path   string `json:"path,omitempty"`   // file path or URL
	Format string `json:"format,omitempty"` // "jsonl" | "json" | "parquet"
}

// GoldConfig defines how gold/expected output is defined.
type GoldConfig struct {
	Type string `json:"type"`           // "exact_match" | "contains" | "regex" | "llm_judge"
	Data any    `json:"data,omitempty"` // file path or embedded data
}

// Threshold defines pass/regression thresholds for scoring.
type Threshold struct {
	Pass            float64 `json:"pass"`             // >= pass = 通过
	RegressionAlert float64 `json:"regression_alert"` // < regression_alert = 回归告警
}

// ScoringConfig defines how a task is scored.
type ScoringConfig struct {
	PrimaryMetric    string    `json:"primary_metric"` // "exact_match" | "semantic_similarity" | "llm_judge"
	SecondaryMetrics []string  `json:"secondary_metrics,omitempty"`
	Threshold        Threshold `json:"threshold"`
}

// ExecutionConfig defines execution parameters for running this task.
type ExecutionConfig struct {
	Model       string  `json:"model"`          // e.g. "gpt-4o" | "claude-3.5" | "llama3"
	Temperature float64 `json:"temperature"`    // 0.0 ~ 2.0
	MaxTokens   int     `json:"max_tokens"`     // 1 ~ 128000
	Seed        int64   `json:"seed,omitempty"` // for deterministic replay
}

// Metric represents a metric definition.
type Metric struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
}

// Evaluation represents an evaluation result.
type Evaluation struct {
	ID              string         `json:"id"`
	TaskExecutionID string         `json:"task_execution_id"`
	MetricID        string         `json:"metric_id"`
	Score           float64        `json:"score"`
	Details         map[string]any `json:"details"`
	EvaluatedAt     time.Time      `json:"evaluated_at"`
}

// TaskExecution represents a task execution record.
type TaskExecution struct {
	ID         string         `json:"id"`
	TaskID     string         `json:"task_id"`
	AgentID    string         `json:"agent_id"`
	Status     string         `json:"status"`
	Input      map[string]any `json:"input"`
	Output     map[string]any `json:"output,omitempty"`
	DurationMs int64          `json:"duration_ms,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// Experiment represents an experiment for comparing agent/skill performance.
type Experiment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TaskIDs     []string  `json:"task_ids"`
	AgentIDs    []string  `json:"agent_ids"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// ExperimentRun represents an individual run within an experiment.
type ExperimentRun struct {
	ID           string         `json:"id"`
	ExperimentID string         `json:"experiment_id"`
	TaskID       string         `json:"task_id"`
	AgentID      string         `json:"agent_id"`
	MetricScores map[string]any `json:"metric_scores"`
	OverallScore float64        `json:"overall_score"`
	DurationMs   int64          `json:"duration_ms,omitempty"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ReplaySnapshot represents a snapshot for deterministic replay.
type ReplaySnapshot struct {
	ID                string         `json:"id"`
	ExecutionID       string         `json:"execution_id"`
	SkillID           string         `json:"skill_id"`
	SkillVersion      string         `json:"skill_version"`
	StepIndex         int            `json:"step_index"`
	StateSnapshot     map[string]any `json:"state_snapshot"`
	InputSeed         map[string]any `json:"input_seed"`
	DeterministicSeed int64          `json:"deterministic_seed"`
	CreatedAt         time.Time      `json:"created_at"`
}

// CreateTaskRequest is the payload for creating a task.
type CreateTaskRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SkillID     string     `json:"skill_id"`
	Tags        []string   `json:"tags"`
	Difficulty  string     `json:"difficulty"`
	TestCases   []TestCase `json:"test_cases"`
}

// CreateMetricRequest is the payload for creating a metric.
type CreateMetricRequest struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

// RunTaskRequest is the payload for running a task evaluation.
type RunTaskRequest struct {
	TaskID  string         `json:"task_id"`
	AgentID string         `json:"agent_id"`
	Input   map[string]any `json:"input"`
}

// EvaluateRequest is the payload for evaluating a task execution.
type EvaluateRequest struct {
	MetricID string `json:"metric_id"`
}

// CreateExperimentRequest is the payload for creating an experiment.
type CreateExperimentRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TaskIDs     []string `json:"task_ids"`
	AgentIDs    []string `json:"agent_ids"`
}
