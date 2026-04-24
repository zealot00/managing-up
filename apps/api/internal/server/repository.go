package server

import (
	"context"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

// Repository defines the persistence contract used by HTTP handlers.
type Repository interface {
	ListSkills(status string) []Skill
	GetSkill(id string) (Skill, bool)
	CreateSkill(req CreateSkillRequest) Skill
	ListDependencies(ctx context.Context, skillID string) ([]SkillDependency, error)
	UpsertRating(ctx context.Context, skillID, userID string, rating int, comment string) error
	ListSkillsByCategory(ctx context.Context, category, search string) ([]Skill, error)
	GetRatingStats(ctx context.Context, skillID string) (float64, int, error)
	GetInstallCount(ctx context.Context, skillID string) (int, error)
	ResolveDepTree(ctx context.Context, skillID string) ([]DependencyNode, error)
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
	GetUserByUsername(username string) (models.User, bool)
	GetUserByID(id string) (models.User, bool)
	CreateUser(user models.User) error
	CreateGatewayAPIKey(key GatewayAPIKey) error
	ListGatewayAPIKeys(userID string) []GatewayAPIKey
	GetGatewayAPIKeyByHash(keyHash string) (GatewayAPIKey, bool)
	TouchGatewayAPIKeyLastUsed(id string, usedAt time.Time) error
	RevokeGatewayAPIKey(id string, userID string) error
	CreateGatewayUsageEvent(event GatewayUsageEvent) error
	ListGatewayUsageByUser(userID string, from, to *time.Time) []GatewayUsageAggregate
	ListGatewayUsageByUsers(from, to *time.Time) []GatewayUserUsageAggregate
	CreateGatewayProviderKey(key GatewayProviderKey) error
	GetGatewayProviderKeyByUserAndProvider(userID, provider string) (GatewayProviderKey, bool)
	ListGatewayProviderKeys(userID string) []GatewayProviderKey
	GetGatewayProviderKey(id string) (GatewayProviderKey, bool)
	GetGatewayProviderKeyByHash(keyHash string) (GatewayProviderKey, bool)
	UpdateGatewayProviderKey(key GatewayProviderKey) error
	DeleteGatewayProviderKey(id string, userID string) error
	ToggleGatewayProviderKey(id string, userID string, enabled bool) error
	GetUserBudget(userID string) (UserBudget, bool)
	CreateOrUpdateUserBudget(budget UserBudget) error
	DecrementUserBudget(userID string, tokens int) (int, error)
	GetRandomTip() (Tip, bool)
	ListMCPServers() []MCPServer
	GetMCPServer(id string) (MCPServer, bool)
	CreateMCPServer(server MCPServer) (MCPServer, error)
	UpdateMCPServer(server MCPServer) error
	DeleteMCPServer(id string) error
	GetLatestSnapshot(ctx context.Context, skillID, version string) (*SkillCapabilitySnapshot, error)
	ListSnapshots(ctx context.Context, skillID string, limit int) ([]SkillCapabilitySnapshot, error)
	ListBridgeAdapterConfigs() []BridgeAdapterConfig
	GetBridgeAdapterConfig(id string) (BridgeAdapterConfig, bool)
	CreateBridgeAdapterConfig(config BridgeAdapterConfig) (BridgeAdapterConfig, error)
	UpdateBridgeAdapterConfig(config BridgeAdapterConfig) error
	DeleteBridgeAdapterConfig(id string) error
	ListPolicyVersions() ([]models.PolicyVersion, error)
	GetPolicyVersion(name string) (models.PolicyVersion, bool)
	CreatePolicyVersion(pv models.PolicyVersion) (models.PolicyVersion, error)
	UpdatePolicyVersion(pv models.PolicyVersion) error
	DeletePolicyVersion(id string) error
	ListMCPServerPermissions(mcpServerID string) ([]MCPServerPermission, error)
	CreateMCPServerPermission(p MCPServerPermission) (MCPServerPermission, error)
	CheckMCPPermission(mcpServerID, userID, apiKeyID, skillID string) (bool, error)
	ListMCPRouterCatalog() ([]MCPRouterCatalogEntry, error)
	UpsertMCPRouterCatalogEntry(e MCPRouterCatalogEntry) error
	IncrementMCPRouterCatalogUseCount(serverID string) error

	// Sweep-related methods
	CreateSweepConfig(cfg SweepConfig) (SweepConfig, error)
	GetSweepConfig(id string) (SweepConfig, bool)
	ListSweepConfigs() ([]SweepConfig, error)
	UpdateSweepConfig(cfg SweepConfig) error
	DeleteSweepConfig(id string) error
	CreateSweepRuns(runs []SweepRun) error
	GetSweepRunsByConfigID(configID string) ([]SweepRun, error)
	UpdateSweepRun(run SweepRun) error
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

type DependencyNode struct {
	SkillID  string           `json:"skill_id"`
	Name     string           `json:"name"`
	Version  string           `json:"version"`
	Children []DependencyNode `json:"children,omitempty"`
}
