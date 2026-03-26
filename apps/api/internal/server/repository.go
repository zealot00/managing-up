package server

import "time"

// Repository defines the persistence contract used by HTTP handlers.
type Repository interface {
	ListSkills(status string) []Skill
	GetSkill(id string) (Skill, bool)
	CreateSkill(req CreateSkillRequest) Skill
	ListSkillVersions(skillID string) []SkillVersion
	GetSkillVersionForExecution(skillID string) (SkillVersion, bool)
	ListProcedureDrafts(status string) []ProcedureDraft
	ListExecutions(status string) []Execution
	GetExecution(id string) (Execution, bool)
	CreateExecution(req CreateExecutionRequest) (Execution, bool)
	ListApprovals(status string) []Approval
	ApproveExecution(executionID string, req ApproveExecutionRequest) (Approval, bool)
	Dashboard() DashboardData
	ListTraces(executionID string) []TraceEvent
	CreateTrace(event TraceEvent) error
	ListTasks(skillID string, difficulty string) []Task
	GetTask(id string) (Task, bool)
	CreateTask(task Task) (Task, error)
	UpdateTask(task Task) error
	DeleteTask(id string) error
	ListMetrics() []Metric
	CreateMetric(metric Metric) (Metric, error)
	GetMetric(id string) (Metric, bool)
	ListTaskExecutions() []TaskExecution
	GetTaskExecution(id string) (TaskExecution, bool)
	CreateTaskExecution(ex TaskExecution) (TaskExecution, error)
	UpdateTaskExecution(ex TaskExecution) error
	ListEvaluations(taskExecutionID string) []Evaluation
	CreateEvaluationResult(eval Evaluation) (Evaluation, error)
	ListExperiments() []Experiment
	GetExperiment(id string) (Experiment, bool)
	CreateExperiment(exp Experiment) (Experiment, error)
	CreateExperimentRun(run ExperimentRun) (ExperimentRun, error)
	UpdateExperimentRun(run ExperimentRun) error
	ListExperimentRuns(experimentID string) []ExperimentRun
	ListReplaySnapshots(executionID string) []ReplaySnapshot
	GetReplaySnapshot(id string) (ReplaySnapshot, bool)
	CreateReplaySnapshot(snap ReplaySnapshot) (ReplaySnapshot, error)
}

// ExecutionRepository extends Repository with methods needed by the runtime engine.
type ExecutionRepository interface {
	Repository
	GetSkillVersionForExecution(skillID string) (SkillVersion, bool)
	UpdateExecutionStatus(id string, status string, stepID string, endedAt *time.Time, durationMs *int64) error
	CreateExecutionStep(step ExecutionStep) error
	GetExecutionForResume(id string) (Execution, bool)
}

// WorkerRepository provides methods for the background worker.
type WorkerRepository interface {
	ListPendingExecutions() []Execution
	ListWaitingApprovalExecutions() []Execution
}

// ExecutionStep represents a step execution record.
type ExecutionStep struct {
	ID          string         `json:"id"`
	ExecutionID string         `json:"execution_id"`
	StepID      string         `json:"step_id"`
	Status      string         `json:"status"`
	ToolRef     string         `json:"tool_ref,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	EndedAt     *time.Time     `json:"ended_at,omitempty"`
	DurationMs  int64          `json:"duration_ms,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	AttemptNo   int            `json:"attempt_no"`
}
