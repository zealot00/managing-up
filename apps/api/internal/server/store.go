package server

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

type store struct {
	mu              sync.RWMutex
	skills          []Skill
	skillVersions   []SkillVersion
	procedureDrafts []ProcedureDraft
	executions      []Execution
	approvals       []Approval
}

var _ Repository = (*store)(nil)

func newStore() *store {
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
	return []Task{}
}

func (s *store) GetTask(id string) (Task, bool) {
	return Task{}, false
}

func (s *store) CreateTask(task Task) (Task, error) {
	return Task{}, nil
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
	return []Experiment{}
}

func (s *store) GetExperiment(id string) (Experiment, bool) {
	return Experiment{}, false
}

func (s *store) CreateExperiment(exp Experiment) (Experiment, error) {
	return Experiment{}, nil
}

func (s *store) ListExperimentRuns(experimentID string) []ExperimentRun {
	return []ExperimentRun{}
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
