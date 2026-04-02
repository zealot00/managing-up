package server

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
)

type store struct {
	mu              sync.RWMutex
	skills          []Skill
	skillVersions   []SkillVersion
	procedureDrafts []ProcedureDraft
	executions      []Execution
	approvals       []Approval
	tasks           map[string]Task
	experiments     map[string]Experiment
	experimentRuns  map[string]ExperimentRun
	users           map[string]models.User
	gatewayAPIKeys  map[string]GatewayAPIKey
	gatewayUsage    []GatewayUsageEvent
}

var _ Repository = (*store)(nil)

// NewStore creates a new in-memory store.
func NewStore() *store {
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)

	return &store{
		skills: []Skill{
			{
				ID:             "skill_001",
				Name:           "restart_service_skill",
				OwnerTeam:      "platform_team",
				RiskLevel:      "medium",
				Status:         "published",
				CurrentVersion: "v1",
			},
			{
				ID:             "skill_002",
				Name:           "collect_logs_skill",
				OwnerTeam:      "sre_team",
				RiskLevel:      "low",
				Status:         "published",
				CurrentVersion: "v3",
			},
			{
				ID:             "skill_003",
				Name:           "rollback_deployment_skill",
				OwnerTeam:      "platform_team",
				RiskLevel:      "high",
				Status:         "draft",
				CurrentVersion: "",
			},
		},
		skillVersions: []SkillVersion{
			{
				ID:               "version_001",
				SkillID:          "skill_001",
				Version:          "v1",
				Status:           "published",
				ChangeSummary:    "Initial restart automation flow.",
				ApprovalRequired: true,
				CreatedAt:        now.Add(-72 * time.Hour),
			},
			{
				ID:               "version_002",
				SkillID:          "skill_002",
				Version:          "v3",
				Status:           "published",
				ChangeSummary:    "Added export safety checks and retry handling.",
				ApprovalRequired: true,
				CreatedAt:        now.Add(-48 * time.Hour),
			},
			{
				ID:               "version_003",
				SkillID:          "skill_003",
				Version:          "v0-draft",
				Status:           "draft",
				ChangeSummary:    "Rollback flow under review.",
				ApprovalRequired: true,
				CreatedAt:        now.Add(-12 * time.Hour),
			},
		},
		procedureDrafts: []ProcedureDraft{
			{
				ID:               "draft_001",
				ProcedureKey:     "runbook_restart_service",
				Title:            "Restart Service Runbook",
				ValidationStatus: "validated",
				RequiredTools:    []string{"monitor_api", "orchestrator_api"},
				SourceType:       "markdown",
				CreatedAt:        now.Add(-96 * time.Hour),
			},
			{
				ID:               "draft_002",
				ProcedureKey:     "collect_production_logs",
				Title:            "Collect Production Logs",
				ValidationStatus: "draft",
				RequiredTools:    []string{"shell_adapter", "storage_exporter"},
				SourceType:       "pdf",
				CreatedAt:        now.Add(-18 * time.Hour),
			},
		},
		executions: []Execution{
			{
				ID:            "exec_001",
				SkillID:       "skill_001",
				SkillName:     "restart_service_skill",
				Status:        "running",
				TriggeredBy:   "sre_oncall",
				StartedAt:     now,
				CurrentStepID: "verify_health",
			},
			{
				ID:            "exec_002",
				SkillID:       "skill_002",
				SkillName:     "collect_logs_skill",
				Status:        "waiting_approval",
				TriggeredBy:   "ops_manager",
				StartedAt:     now.Add(-35 * time.Minute),
				CurrentStepID: "approval_before_export",
			},
			{
				ID:            "exec_003",
				SkillID:       "skill_001",
				SkillName:     "restart_service_skill",
				Status:        "succeeded",
				TriggeredBy:   "platform_operator",
				StartedAt:     now.Add(-2 * time.Hour),
				CurrentStepID: "completed",
			},
		},
		approvals: []Approval{
			{
				ID:            "approval_001",
				ExecutionID:   "exec_002",
				SkillName:     "collect_logs_skill",
				StepID:        "approval_before_export",
				Status:        "waiting",
				ApproverGroup: "ops_manager",
				RequestedAt:   now.Add(-30 * time.Minute),
			},
		},
		tasks:          make(map[string]Task),
		experiments:    make(map[string]Experiment),
		experimentRuns: make(map[string]ExperimentRun),
		users:          make(map[string]models.User),
		gatewayAPIKeys: make(map[string]GatewayAPIKey),
		gatewayUsage:   make([]GatewayUsageEvent, 0),
	}
}

func (s *store) ListSkills(status string) []Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if status == "" {
		return slices.Clone(s.skills)
	}

	items := make([]Skill, 0, len(s.skills))
	for _, skill := range s.skills {
		if skill.Status == status {
			items = append(items, skill)
		}
	}

	return items
}

func (s *store) GetSkill(id string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, skill := range s.skills {
		if skill.ID == id {
			return skill, true
		}
	}

	return Skill{}, false
}

func (s *store) CreateSkill(req CreateSkillRequest) Skill {
	s.mu.Lock()
	defer s.mu.Unlock()

	skill := Skill{
		ID:             fmt.Sprintf("skill_%03d", len(s.skills)+1),
		Name:           req.Name,
		OwnerTeam:      req.OwnerTeam,
		RiskLevel:      req.RiskLevel,
		Status:         "draft",
		CurrentVersion: "",
	}

	s.skills = append(s.skills, skill)
	return skill
}

func (s *store) ListSkillVersions(skillID string) []SkillVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]SkillVersion, 0, len(s.skillVersions))
	for _, version := range s.skillVersions {
		if skillID == "" || version.SkillID == skillID {
			items = append(items, version)
		}
	}

	return items
}

func (s *store) ListProcedureDrafts(status string) []ProcedureDraft {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if status == "" {
		return slices.Clone(s.procedureDrafts)
	}

	items := make([]ProcedureDraft, 0, len(s.procedureDrafts))
	for _, draft := range s.procedureDrafts {
		if draft.ValidationStatus == status {
			items = append(items, draft)
		}
	}

	return items
}

func (s *store) ListExecutions(status string) []Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if status == "" {
		return slices.Clone(s.executions)
	}

	items := make([]Execution, 0, len(s.executions))
	for _, execution := range s.executions {
		if execution.Status == status {
			items = append(items, execution)
		}
	}

	return items
}

func (s *store) GetExecution(id string) (Execution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, execution := range s.executions {
		if execution.ID == id {
			return execution, true
		}
	}

	return Execution{}, false
}

func (s *store) CreateExecution(req CreateExecutionRequest) (Execution, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var skill Skill
	found := false
	for _, item := range s.skills {
		if item.ID == req.SkillID {
			skill = item
			found = true
			break
		}
	}

	if !found {
		return Execution{}, false
	}

	execution := Execution{
		ID:            fmt.Sprintf("exec_%03d", len(s.executions)+1),
		SkillID:       skill.ID,
		SkillName:     skill.Name,
		Status:        "pending",
		TriggeredBy:   req.TriggeredBy,
		StartedAt:     time.Now().UTC(),
		CurrentStepID: "queued",
		Input:         req.Input,
	}

	s.executions = append([]Execution{execution}, s.executions...)
	return execution, true
}

func (s *store) ListApprovals(status string) []Approval {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if status == "" {
		return slices.Clone(s.approvals)
	}

	items := make([]Approval, 0, len(s.approvals))
	for _, approval := range s.approvals {
		if approval.Status == status {
			items = append(items, approval)
		}
	}

	return items
}

func (s *store) ApproveExecution(executionID string, req ApproveExecutionRequest) (Approval, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, approval := range s.approvals {
		if approval.ExecutionID == executionID {
			s.approvals[i].Status = req.Decision
			s.approvals[i].ApprovedBy = req.Approver
			s.approvals[i].ResolutionNote = req.Note

			for j, execution := range s.executions {
				if execution.ID == executionID {
					switch req.Decision {
					case "approved":
						s.executions[j].Status = "running"
						s.executions[j].CurrentStepID = "resumed_after_approval"
					case "rejected":
						s.executions[j].Status = "failed"
						s.executions[j].CurrentStepID = "approval_rejected"
					}
					break
				}
			}

			return s.approvals[i], true
		}
	}

	return Approval{}, false
}

func (s *store) Dashboard() DashboardData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := DashboardSummary{
		SuccessRate:        0.91,
		AvgDurationSeconds: 84,
	}

	for _, skill := range s.skills {
		if skill.Status != "deprecated" {
			summary.ActiveSkills++
		}
		if skill.CurrentVersion != "" {
			summary.PublishedVersions++
		}
	}

	for _, execution := range s.executions {
		switch execution.Status {
		case "running":
			summary.RunningExecutions++
		case "waiting_approval":
			summary.WaitingApprovals++
		}
	}

	recent := slices.Clone(s.executions)
	if len(recent) > 5 {
		recent = recent[:5]
	}

	return DashboardData{
		Summary:          summary,
		RecentExecutions: recent,
	}
}

func (s *store) GetSkillVersionForExecution(skillID string) (SkillVersion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.skillVersions) - 1; i >= 0; i-- {
		v := s.skillVersions[i]
		if v.SkillID == skillID && v.Status == "published" {
			return v, true
		}
	}

	return SkillVersion{}, false
}

func (s *store) UpdateExecutionStatus(id string, status string, stepID string, endedAt *time.Time, durationMs *int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, execution := range s.executions {
		if execution.ID == id {
			s.executions[i].Status = status
			s.executions[i].CurrentStepID = stepID
			return nil
		}
	}

	return fmt.Errorf("execution not found: %s", id)
}

func (s *store) CreateExecutionStep(step ExecutionStep) error {
	return nil
}

func (s *store) GetExecutionForResume(id string) (Execution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, execution := range s.executions {
		if execution.ID == id {
			if execution.Status == "pending" || execution.Status == "running" || execution.Status == "waiting_approval" {
				return execution, true
			}
		}
	}

	return Execution{}, false
}

func (s *store) ListPendingExecutions() []Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Execution, 0)
	for _, execution := range s.executions {
		if execution.Status == "pending" {
			items = append(items, execution)
		}
	}

	return items
}

func (s *store) ListWaitingApprovalExecutions() []Execution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Execution, 0)
	for _, execution := range s.executions {
		if execution.Status == "waiting_approval" {
			items = append(items, execution)
		}
	}

	return items
}

func (s *store) ListTraces(executionID string) []TraceEvent {
	return []TraceEvent{}
}

func (s *store) CreateTrace(event TraceEvent) error {
	return nil
}

func (s *store) ListTasks(skillID string, difficulty string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if skillID != "" && task.SkillID != skillID {
			continue
		}
		if difficulty != "" && task.Difficulty != difficulty {
			continue
		}
		items = append(items, task)
	}

	return items
}

func (s *store) GetTask(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, found := s.tasks[id]
	return task, found
}

func (s *store) CreateTask(task Task) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%03d", len(s.tasks)+1)
	}
	task.CreatedAt = now
	task.UpdatedAt = now

	s.tasks[task.ID] = task
	return task, nil
}

func (s *store) UpdateTask(task Task) error {
	return nil
}

func (s *store) DeleteTask(id string) error {
	return nil
}

func (s *store) ListMetrics() []Metric {
	return []Metric{}
}

func (s *store) CreateMetric(metric Metric) (Metric, error) {
	return Metric{}, nil
}

func (s *store) GetMetric(id string) (Metric, bool) {
	return Metric{}, false
}

func (s *store) ListTaskExecutions() []TaskExecution {
	return []TaskExecution{}
}

func (s *store) GetTaskExecution(id string) (TaskExecution, bool) {
	return TaskExecution{}, false
}

func (s *store) CreateTaskExecution(ex TaskExecution) (TaskExecution, error) {
	return TaskExecution{}, nil
}

func (s *store) UpdateTaskExecution(ex TaskExecution) error {
	return nil
}

func (s *store) ListEvaluations(taskExecutionID string) []Evaluation {
	return []Evaluation{}
}

func (s *store) CreateEvaluationResult(eval Evaluation) (Evaluation, error) {
	return Evaluation{}, nil
}

func (s *store) ListExperiments() []Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Experiment, 0, len(s.experiments))
	for _, exp := range s.experiments {
		items = append(items, exp)
	}
	return items
}

func (s *store) GetExperiment(id string) (Experiment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, found := s.experiments[id]
	return exp, found
}

func (s *store) CreateExperiment(exp Experiment) (Experiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if exp.ID == "" {
		exp.ID = fmt.Sprintf("exp_%d", time.Now().UnixNano())
	}
	exp.CreatedAt = now
	exp.UpdatedAt = now
	s.experiments[exp.ID] = exp
	return exp, nil
}

func (s *store) CreateExperimentRun(run ExperimentRun) (ExperimentRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if run.ID == "" {
		run.ID = fmt.Sprintf("run_%d", time.Now().UnixNano())
	}
	run.CreatedAt = now
	s.experimentRuns[run.ID] = run
	return run, nil
}

func (s *store) UpdateExperimentRun(run ExperimentRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.experimentRuns[run.ID]; !found {
		return fmt.Errorf("experiment run not found: %s", run.ID)
	}
	s.experimentRuns[run.ID] = run
	return nil
}

func (s *store) ListExperimentRuns(experimentID string) []ExperimentRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ExperimentRun, 0)
	for _, run := range s.experimentRuns {
		if experimentID == "" || run.ExperimentID == experimentID {
			items = append(items, run)
		}
	}
	return items
}

func (s *store) ListReplaySnapshots(executionID string) []ReplaySnapshot {
	return []ReplaySnapshot{}
}

func (s *store) GetReplaySnapshot(id string) (ReplaySnapshot, bool) {
	return ReplaySnapshot{}, false
}

func (s *store) CreateReplaySnapshot(snap ReplaySnapshot) (ReplaySnapshot, error) {
	return ReplaySnapshot{}, nil
}

func (s *store) GetUserByUsername(username string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Username == username {
			return u, true
		}
	}
	return models.User{}, false
}

func (s *store) GetUserByID(id string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *store) CreateGatewayAPIKey(key GatewayAPIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gatewayAPIKeys[key.ID] = key
	return nil
}

func (s *store) ListGatewayAPIKeys(userID string) []GatewayAPIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]GatewayAPIKey, 0)
	for _, item := range s.gatewayAPIKeys {
		if userID == "" || item.UserID == userID {
			items = append(items, item)
		}
	}
	return items
}

func (s *store) GetGatewayAPIKeyByHash(keyHash string) (GatewayAPIKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.gatewayAPIKeys {
		if item.KeyHash == keyHash && item.RevokedAt == nil {
			if user, ok := s.users[item.UserID]; ok {
				item.Username = user.Username
				item.Role = user.Role
			}
			return item, true
		}
	}
	return GatewayAPIKey{}, false
}

func (s *store) TouchGatewayAPIKeyLastUsed(id string, usedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.gatewayAPIKeys[id]
	if !ok {
		return fmt.Errorf("gateway api key not found: %s", id)
	}
	item.LastUsedAt = &usedAt
	s.gatewayAPIKeys[id] = item
	return nil
}

func (s *store) RevokeGatewayAPIKey(id string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.gatewayAPIKeys[id]
	if !ok {
		return fmt.Errorf("gateway api key not found: %s", id)
	}
	if userID != "" && item.UserID != userID {
		return fmt.Errorf("gateway api key does not belong to user: %s", userID)
	}
	now := time.Now().UTC()
	item.RevokedAt = &now
	s.gatewayAPIKeys[id] = item
	return nil
}

func (s *store) CreateGatewayUsageEvent(event GatewayUsageEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gatewayUsage = append(s.gatewayUsage, event)
	return nil
}

func (s *store) ListGatewayUsageByUser(userID string, from, to *time.Time) []GatewayUsageAggregate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc := map[string]GatewayUsageAggregate{}
	for _, item := range s.gatewayUsage {
		if userID != "" && item.UserID != userID {
			continue
		}
		if from != nil && item.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && item.CreatedAt.After(*to) {
			continue
		}
		key := item.Provider + "|" + item.Model
		agg := acc[key]
		agg.UserID = item.UserID
		agg.Provider = item.Provider
		agg.Model = item.Model
		agg.RequestCount++
		agg.PromptTokens += int64(item.PromptTokens)
		agg.CompletionTokens += int64(item.CompletionTokens)
		agg.TotalTokens += int64(item.TotalTokens)
		agg.TotalCost += item.Cost
		if agg.Username == "" {
			if user, ok := s.users[item.UserID]; ok {
				agg.Username = user.Username
			}
		}
		acc[key] = agg
	}

	items := make([]GatewayUsageAggregate, 0, len(acc))
	for _, item := range acc {
		items = append(items, item)
	}
	return items
}

func (s *store) ListGatewayUsageByUsers(from, to *time.Time) []GatewayUserUsageAggregate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc := map[string]GatewayUserUsageAggregate{}
	for _, item := range s.gatewayUsage {
		if from != nil && item.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && item.CreatedAt.After(*to) {
			continue
		}
		agg := acc[item.UserID]
		agg.UserID = item.UserID
		agg.RequestCount++
		agg.PromptTokens += int64(item.PromptTokens)
		agg.CompletionTokens += int64(item.CompletionTokens)
		agg.TotalTokens += int64(item.TotalTokens)
		agg.TotalCost += item.Cost
		if user, ok := s.users[item.UserID]; ok {
			agg.Username = user.Username
		}
		acc[item.UserID] = agg
	}

	items := make([]GatewayUserUsageAggregate, 0, len(acc))
	for _, item := range acc {
		items = append(items, item)
	}
	return items
}

func (s *store) CreateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
	return nil
}

func (s *store) GetRandomTip() (Tip, bool) {
	return Tip{}, false
}
