package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"github.com/zealot/managing-up/apps/api/internal/server"
)

// Repository implements the server.Repository contract with PostgreSQL storage.
type Repository struct {
	db *sql.DB
}

// New opens a PostgreSQL-backed repository and verifies connectivity.
func New(dsn string) (*Repository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	// Ensure new columns exist ( idempotent ALTER TABLE )
	if _, err := db.Exec(`ALTER TABLE experiment_runs ADD COLUMN IF NOT EXISTS variant_id TEXT NOT NULL DEFAULT ''`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate experiment_runs variant_id: %w", err)
	}
	if _, err := db.Exec(`ALTER TABLE experiments ADD COLUMN IF NOT EXISTS variants JSONB DEFAULT '[]'::jsonb`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate experiments variants: %w", err)
	}

	// Gateway provider keys table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gateway_provider_keys (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			model TEXT DEFAULT '',
			key_hash TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			encrypted_key TEXT NOT NULL,
			is_enabled BOOLEAN DEFAULT true,
			monthly_limit INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)
	`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate gateway_provider_keys: %w", err)
	}

	// User budgets table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_budgets (
			id TEXT PRIMARY KEY,
			user_id TEXT UNIQUE NOT NULL,
			monthly_limit INTEGER DEFAULT 0,
			daily_limit INTEGER DEFAULT 0,
			used_this_month INTEGER DEFAULT 0,
			used_today INTEGER DEFAULT 0,
			reset_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)
	`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate user_budgets: %w", err)
	}

	return &Repository{db: db}, nil
}

// Close releases the database handle.
func (r *Repository) Close() error {
	return r.db.Close()
}

// ListSkills returns skill metadata records.
func (r *Repository) ListSkills(status string) []server.Skill {
	query := `
		SELECT id, name, owner_team, risk_level, status, current_version, COALESCE(created_by, ''), updated_at
		FROM skills
		WHERE ($1 = '' OR status = $1)
		ORDER BY name
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return []server.Skill{}
	}
	defer rows.Close()

	items := make([]server.Skill, 0)
	for rows.Next() {
		var item server.Skill
		if err := rows.Scan(&item.ID, &item.Name, &item.OwnerTeam, &item.RiskLevel, &item.Status, &item.CurrentVersion, &item.CreatedBy, &item.UpdatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items
}

// GetSkill returns a single skill record.
func (r *Repository) GetSkill(id string) (server.Skill, bool) {
	query := `
		SELECT id, name, owner_team, risk_level, status, current_version, COALESCE(created_by, ''), updated_at
		FROM skills
		WHERE id = $1
	`

	var item server.Skill
	err := r.db.QueryRow(query, id).Scan(&item.ID, &item.Name, &item.OwnerTeam, &item.RiskLevel, &item.Status, &item.CurrentVersion, &item.CreatedBy, &item.UpdatedAt)
	if err != nil {
		return server.Skill{}, false
	}

	return item, true
}

// CreateSkill inserts a draft skill record.
func (r *Repository) CreateSkill(req server.CreateSkillRequest) server.Skill {
	id := fmt.Sprintf("skill_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	item := server.Skill{
		ID:             id,
		Name:           req.Name,
		OwnerTeam:      req.OwnerTeam,
		RiskLevel:      req.RiskLevel,
		Status:         "draft",
		CurrentVersion: "",
		CreatedBy:      req.OwnerTeam,
		UpdatedAt:      now,
	}

	query := `
		INSERT INTO skills (id, name, owner_team, risk_level, status, current_version, created_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if _, err := r.db.Exec(query, item.ID, item.Name, item.OwnerTeam, item.RiskLevel, item.Status, item.CurrentVersion, item.CreatedBy, item.UpdatedAt); err != nil {
		return server.Skill{}
	}

	return item
}

// ListSkillVersions returns version rows.
func (r *Repository) ListSkillVersions(skillID string) []server.SkillVersion {
	query := `
		SELECT id, skill_id, version, status, change_summary, approval_required, created_at
		FROM skill_versions
		WHERE ($1 = '' OR skill_id = $1)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, skillID)
	if err != nil {
		return []server.SkillVersion{}
	}
	defer rows.Close()

	items := make([]server.SkillVersion, 0)
	for rows.Next() {
		var item server.SkillVersion
		if err := rows.Scan(&item.ID, &item.SkillID, &item.Version, &item.Status, &item.ChangeSummary, &item.ApprovalRequired, &item.CreatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items
}

// ListProcedureDrafts returns procedure draft rows.
func (r *Repository) ListProcedureDrafts(status string) []server.ProcedureDraft {
	query := `
		SELECT id, procedure_key, title, validation_status, required_tools, source_type, created_at
		FROM procedure_drafts
		WHERE ($1 = '' OR validation_status = $1)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return []server.ProcedureDraft{}
	}
	defer rows.Close()

	items := make([]server.ProcedureDraft, 0)
	for rows.Next() {
		var item server.ProcedureDraft
		var rawTools []byte
		if err := rows.Scan(&item.ID, &item.ProcedureKey, &item.Title, &item.ValidationStatus, &rawTools, &item.SourceType, &item.CreatedAt); err != nil {
			continue
		}
		if err := json.Unmarshal(rawTools, &item.RequiredTools); err != nil {
			item.RequiredTools = []string{}
		}
		items = append(items, item)
	}

	return items
}

// ListExecutions returns execution rows.
func (r *Repository) ListExecutions(status string) []server.Execution {
	query := `
		SELECT id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, COALESCE(created_by, '')
		FROM executions
		WHERE ($1 = '' OR status = $1)
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return []server.Execution{}
	}
	defer rows.Close()

	items := make([]server.Execution, 0)
	for rows.Next() {
		var item server.Execution
		var rawInput []byte
		if err := rows.Scan(&item.ID, &item.SkillID, &item.SkillName, &item.Status, &item.TriggeredBy, &item.StartedAt, &item.CurrentStepID, &rawInput, &item.CreatedBy); err != nil {
			continue
		}
		if err := json.Unmarshal(rawInput, &item.Input); err != nil {
			item.Input = map[string]any{}
		}
		items = append(items, item)
	}

	return items
}

// GetExecution returns one execution row.
func (r *Repository) GetExecution(id string) (server.Execution, bool) {
	query := `
		SELECT id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, COALESCE(created_by, '')
		FROM executions
		WHERE id = $1
	`

	var item server.Execution
	var rawInput []byte
	err := r.db.QueryRow(query, id).Scan(&item.ID, &item.SkillID, &item.SkillName, &item.Status, &item.TriggeredBy, &item.StartedAt, &item.CurrentStepID, &rawInput, &item.CreatedBy)
	if err != nil {
		return server.Execution{}, false
	}

	if err := json.Unmarshal(rawInput, &item.Input); err != nil {
		item.Input = map[string]any{}
	}

	return item, true
}

// CreateExecution inserts a new execution row for an existing skill.
func (r *Repository) CreateExecution(req server.CreateExecutionRequest) (server.Execution, bool) {
	skill, ok := r.GetSkill(req.SkillID)
	if !ok {
		return server.Execution{}, false
	}

	id := fmt.Sprintf("exec_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	rawInput, err := json.Marshal(req.Input)
	if err != nil {
		return server.Execution{}, false
	}

	item := server.Execution{
		ID:            id,
		SkillID:       skill.ID,
		SkillName:     skill.Name,
		Status:        "pending",
		TriggeredBy:   req.TriggeredBy,
		StartedAt:     now,
		CurrentStepID: "queued",
		Input:         req.Input,
		CreatedBy:     req.TriggeredBy,
	}

	query := `
		INSERT INTO executions (id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if _, err := r.db.Exec(query, item.ID, item.SkillID, item.SkillName, item.Status, item.TriggeredBy, item.StartedAt, item.CurrentStepID, rawInput, item.CreatedBy); err != nil {
		return server.Execution{}, false
	}

	return item, true
}

// ListApprovals returns approval rows.
func (r *Repository) ListApprovals(status string) []server.Approval {
	query := `
		SELECT id, execution_id, skill_name, step_id, status, approver_group, requested_at, COALESCE(approved_by, ''), COALESCE(resolution_note, '')
		FROM approvals
		WHERE ($1 = '' OR status = $1)
		ORDER BY requested_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return []server.Approval{}
	}
	defer rows.Close()

	items := make([]server.Approval, 0)
	for rows.Next() {
		var item server.Approval
		if err := rows.Scan(&item.ID, &item.ExecutionID, &item.SkillName, &item.StepID, &item.Status, &item.ApproverGroup, &item.RequestedAt, &item.ApprovedBy, &item.ResolutionNote); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items
}

// ApproveExecution updates approval and execution state.
func (r *Repository) ApproveExecution(executionID string, req server.ApproveExecutionRequest) (server.Approval, bool) {
	tx, err := r.db.Begin()
	if err != nil {
		return server.Approval{}, false
	}
	defer tx.Rollback()

	status := "running"
	stepID := "resumed_after_approval"
	if req.Decision == "rejected" {
		status = "failed"
		stepID = "approval_rejected"
	}

	if _, err := tx.Exec(`UPDATE approvals SET status = $1, approved_by = $2, resolution_note = $3 WHERE execution_id = $4`, req.Decision, req.Approver, req.Note, executionID); err != nil {
		return server.Approval{}, false
	}

	if _, err := tx.Exec(`UPDATE executions SET status = $1, current_step_id = $2 WHERE id = $3`, status, stepID, executionID); err != nil {
		return server.Approval{}, false
	}

	var item server.Approval
	err = tx.QueryRow(`
		SELECT id, execution_id, skill_name, step_id, status, approver_group, requested_at, COALESCE(approved_by, ''), COALESCE(resolution_note, '')
		FROM approvals
		WHERE execution_id = $1
	`, executionID).Scan(&item.ID, &item.ExecutionID, &item.SkillName, &item.StepID, &item.Status, &item.ApproverGroup, &item.RequestedAt, &item.ApprovedBy, &item.ResolutionNote)
	if err != nil {
		return server.Approval{}, false
	}

	if err := tx.Commit(); err != nil {
		return server.Approval{}, false
	}

	return item, true
}

// Dashboard returns homepage metrics using current repository data.
func (r *Repository) GetSkillVersionForExecution(skillID string) (server.SkillVersion, bool) {
	query := `
		SELECT sv.id, sv.skill_id, sv.version, sv.status, sv.change_summary, sv.approval_required, sv.created_at, COALESCE(sv.spec_yaml, '')
		FROM skill_versions sv
		JOIN skills s ON s.id = sv.skill_id
		WHERE sv.skill_id = $1 AND sv.status = 'published'
		ORDER BY sv.created_at DESC
		LIMIT 1
	`

	var item server.SkillVersion
	var specYaml string
	err := r.db.QueryRow(query, skillID).Scan(&item.ID, &item.SkillID, &item.Version, &item.Status, &item.ChangeSummary, &item.ApprovalRequired, &item.CreatedAt, &specYaml)
	if err != nil {
		return server.SkillVersion{}, false
	}

	return item, true
}

func (r *Repository) UpdateExecutionStatus(id string, status string, stepID string, endedAt *time.Time, durationMs *int64) error {
	query := `
		UPDATE executions 
		SET status = $2, current_step_id = $3, ended_at = $4, duration_ms = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, status, stepID, endedAt, durationMs)
	return err
}

func (r *Repository) CreateExecutionStep(step server.ExecutionStep) error {
	query := `
		INSERT INTO execution_steps (id, execution_id, step_id, status, tool_ref, started_at, ended_at, duration_ms, output, error, attempt_no)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	var outputJSON []byte
	var err error
	if step.Output != nil {
		outputJSON, err = json.Marshal(step.Output)
		if err != nil {
			outputJSON = []byte("{}")
		}
	} else {
		outputJSON = []byte("{}")
	}

	_, err = r.db.Exec(query, step.ID, step.ExecutionID, step.StepID, step.Status, step.ToolRef, step.StartedAt, step.EndedAt, step.DurationMs, outputJSON, step.Error, step.AttemptNo)
	return err
}

func (r *Repository) GetExecutionForResume(id string) (server.Execution, bool) {
	query := `
		SELECT id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, COALESCE(created_by, '')
		FROM executions
		WHERE id = $1 AND status IN ('pending', 'running', 'waiting_approval')
	`

	var item server.Execution
	var rawInput []byte
	err := r.db.QueryRow(query, id).Scan(&item.ID, &item.SkillID, &item.SkillName, &item.Status, &item.TriggeredBy, &item.StartedAt, &item.CurrentStepID, &rawInput, &item.CreatedBy)
	if err != nil {
		return server.Execution{}, false
	}

	if err := json.Unmarshal(rawInput, &item.Input); err != nil {
		item.Input = map[string]any{}
	}

	return item, true
}

func (r *Repository) ListPendingExecutions() []server.Execution {
	query := `
		SELECT id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, COALESCE(created_by, '')
		FROM executions
		WHERE status = 'pending'
		ORDER BY started_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return []server.Execution{}
	}
	defer rows.Close()

	items := make([]server.Execution, 0)
	for rows.Next() {
		var item server.Execution
		var rawInput []byte
		if err := rows.Scan(&item.ID, &item.SkillID, &item.SkillName, &item.Status, &item.TriggeredBy, &item.StartedAt, &item.CurrentStepID, &rawInput, &item.CreatedBy); err != nil {
			continue
		}
		if err := json.Unmarshal(rawInput, &item.Input); err != nil {
			item.Input = map[string]any{}
		}
		items = append(items, item)
	}

	return items
}

func (r *Repository) ListWaitingApprovalExecutions() []server.Execution {
	query := `
		SELECT id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input, COALESCE(created_by, '')
		FROM executions
		WHERE status = 'waiting_approval'
		ORDER BY started_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return []server.Execution{}
	}
	defer rows.Close()

	items := make([]server.Execution, 0)
	for rows.Next() {
		var item server.Execution
		var rawInput []byte
		if err := rows.Scan(&item.ID, &item.SkillID, &item.SkillName, &item.Status, &item.TriggeredBy, &item.StartedAt, &item.CurrentStepID, &rawInput, &item.CreatedBy); err != nil {
			continue
		}
		if err := json.Unmarshal(rawInput, &item.Input); err != nil {
			item.Input = map[string]any{}
		}
		items = append(items, item)
	}

	return items
}

func (r *Repository) Dashboard() server.DashboardData {
	skills := r.ListSkills("")
	executions := r.ListExecutions("")

	summary := server.DashboardSummary{
		SuccessRate:        0.91,
		AvgDurationSeconds: 84,
	}

	for _, skill := range skills {
		if skill.Status != "deprecated" {
			summary.ActiveSkills++
		}
		if skill.CurrentVersion != "" {
			summary.PublishedVersions++
		}
	}

	for _, execution := range executions {
		switch execution.Status {
		case "running":
			summary.RunningExecutions++
		case "waiting_approval":
			summary.WaitingApprovals++
		}
	}

	if len(executions) > 5 {
		executions = executions[:5]
	}

	return server.DashboardData{
		Summary:          summary,
		RecentExecutions: executions,
	}
}

func (r *Repository) CreateTrace(event server.TraceEvent) error {
	query := `
		INSERT INTO execution_traces (id, execution_id, step_id, event_type, event_data, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		eventDataJSON = []byte("{}")
	}
	_, err = r.db.Exec(query, event.ID, event.ExecutionID, event.StepID, event.EventType, eventDataJSON, event.Timestamp)
	return err
}

func (r *Repository) ListTraces(executionID string) []server.TraceEvent {
	query := `
		SELECT id, execution_id, step_id, event_type, event_data, timestamp
		FROM execution_traces
		WHERE execution_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(query, executionID)
	if err != nil {
		return []server.TraceEvent{}
	}
	defer rows.Close()

	items := make([]server.TraceEvent, 0)
	for rows.Next() {
		var item server.TraceEvent
		var stepID sql.NullString
		var eventDataJSON []byte
		if err := rows.Scan(&item.ID, &item.ExecutionID, &stepID, &item.EventType, &eventDataJSON, &item.Timestamp); err != nil {
			continue
		}
		if stepID.Valid {
			item.StepID = stepID.String
		}
		if err := json.Unmarshal(eventDataJSON, &item.EventData); err != nil {
			item.EventData = map[string]any{}
		}
		items = append(items, item)
	}

	return items
}

func (r *Repository) CreateTask(task server.Task) (server.Task, error) {
	id := task.ID
	if id == "" {
		id = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// Default values for new columns
	taskType := task.TaskType
	if taskType == "" {
		taskType = "benchmark"
	}
	inputSource := task.Input.Source
	if inputSource == "" {
		inputSource = "inline"
	}
	goldType := task.Gold.Type
	if goldType == "" {
		goldType = "exact_match"
	}
	primaryMetric := task.Scoring.PrimaryMetric
	if primaryMetric == "" {
		primaryMetric = "exact_match"
	}
	executionModel := task.Execution.Model
	if executionModel == "" {
		executionModel = "gpt-4o"
	}

	tagsJSON, _ := json.Marshal(task.Tags)
	testCasesJSON, _ := json.Marshal(task.TestCases)
	goldDataJSON, _ := json.Marshal(task.Gold.Data)
	secondaryMetricsJSON, _ := json.Marshal(task.Scoring.SecondaryMetrics)

	query := `
		INSERT INTO tasks (
			id, name, description, skill_id, tags, difficulty, test_cases,
			created_at, updated_at,
			task_type, input_source, input_path, input_format,
			gold_type, gold_data, primary_metric, secondary_metrics,
			threshold_pass, threshold_regression_alert,
			execution_model, execution_temperature, execution_max_tokens, execution_seed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`
	// Use nil for empty skill_id to satisfy FK constraint
	var skillIDArg any
	if task.SkillID != "" {
		skillIDArg = task.SkillID
	}

	_, err := r.db.Exec(query,
		id, task.Name, task.Description, skillIDArg, tagsJSON, task.Difficulty, testCasesJSON,
		task.CreatedAt, task.UpdatedAt,
		taskType, inputSource, task.Input.Path, task.Input.Format,
		goldType, goldDataJSON, primaryMetric, secondaryMetricsJSON,
		task.Scoring.Threshold.Pass, task.Scoring.Threshold.RegressionAlert,
		executionModel, task.Execution.Temperature, task.Execution.MaxTokens, task.Execution.Seed,
	)
	if err != nil {
		return server.Task{}, err
	}

	task.ID = id
	return task, nil
}

func (r *Repository) GetTask(id string) (server.Task, bool) {
	query := `
		SELECT
			id, name, description, skill_id, tags, difficulty, test_cases,
			created_at, updated_at,
			task_type, input_source, input_path, input_format,
			gold_type, gold_data, primary_metric, secondary_metrics,
			threshold_pass, threshold_regression_alert,
			execution_model, execution_temperature, execution_max_tokens, execution_seed
		FROM tasks
		WHERE id = $1
	`

	var task server.Task
	var tagsJSON, testCasesJSON, goldDataJSON, secondaryMetricsJSON []byte
	var skillID *string

	err := r.db.QueryRow(query, id).Scan(
		&task.ID, &task.Name, &task.Description, &skillID,
		&tagsJSON, &task.Difficulty, &testCasesJSON,
		&task.CreatedAt, &task.UpdatedAt,
		&task.TaskType, &task.Input.Source, &task.Input.Path, &task.Input.Format,
		&task.Gold.Type, &goldDataJSON, &task.Scoring.PrimaryMetric, &secondaryMetricsJSON,
		&task.Scoring.Threshold.Pass, &task.Scoring.Threshold.RegressionAlert,
		&task.Execution.Model, &task.Execution.Temperature, &task.Execution.MaxTokens, &task.Execution.Seed,
	)
	if err != nil {
		return server.Task{}, false
	}

	if skillID != nil {
		task.SkillID = *skillID
	} else {
		task.SkillID = ""
	}
	json.Unmarshal(tagsJSON, &task.Tags)
	json.Unmarshal(testCasesJSON, &task.TestCases)
	json.Unmarshal(goldDataJSON, &task.Gold.Data)
	json.Unmarshal(secondaryMetricsJSON, &task.Scoring.SecondaryMetrics)

	return task, true
}

func (r *Repository) ListTasks(skillID string, difficulty string) []server.Task {
	query := `
		SELECT
			id, name, description, skill_id, tags, difficulty, test_cases,
			created_at, updated_at,
			task_type, input_source, input_path, input_format,
			gold_type, gold_data, primary_metric, secondary_metrics,
			threshold_pass, threshold_regression_alert,
			execution_model, execution_temperature, execution_max_tokens, execution_seed
		FROM tasks
		WHERE ($1 = '' OR skill_id = $1) AND ($2 = '' OR difficulty = $2)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, skillID, difficulty)
	if err != nil {
		return []server.Task{}
	}
	defer rows.Close()

	items := make([]server.Task, 0)
	for rows.Next() {
		var task server.Task
		var tagsJSON, testCasesJSON, goldDataJSON, secondaryMetricsJSON []byte
		var sID *string

		if err := rows.Scan(
			&task.ID, &task.Name, &task.Description, &sID,
			&tagsJSON, &task.Difficulty, &testCasesJSON,
			&task.CreatedAt, &task.UpdatedAt,
			&task.TaskType, &task.Input.Source, &task.Input.Path, &task.Input.Format,
			&task.Gold.Type, &goldDataJSON, &task.Scoring.PrimaryMetric, &secondaryMetricsJSON,
			&task.Scoring.Threshold.Pass, &task.Scoring.Threshold.RegressionAlert,
			&task.Execution.Model, &task.Execution.Temperature, &task.Execution.MaxTokens, &task.Execution.Seed,
		); err != nil {
			continue
		}

		if sID != nil {
			task.SkillID = *sID
		} else {
			task.SkillID = ""
		}
		json.Unmarshal(tagsJSON, &task.Tags)
		json.Unmarshal(testCasesJSON, &task.TestCases)
		json.Unmarshal(goldDataJSON, &task.Gold.Data)
		json.Unmarshal(secondaryMetricsJSON, &task.Scoring.SecondaryMetrics)
		items = append(items, task)
	}

	return items
}

func (r *Repository) UpdateTask(task server.Task) error {
	tagsJSON, _ := json.Marshal(task.Tags)
	testCasesJSON, _ := json.Marshal(task.TestCases)

	query := `
		UPDATE tasks
		SET name = $2, description = $3, skill_id = $4, tags = $5, difficulty = $6, test_cases = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(query, task.ID, task.Name, task.Description, task.SkillID, tagsJSON, task.Difficulty, testCasesJSON, task.UpdatedAt)
	return err
}

func (r *Repository) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *Repository) CreateMetric(metric server.Metric) (server.Metric, error) {
	id := metric.ID
	if id == "" {
		id = fmt.Sprintf("metric_%d", time.Now().UnixNano())
	}

	configJSON, _ := json.Marshal(metric.Config)

	query := `
		INSERT INTO metrics (id, name, type, config, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, id, metric.Name, metric.Type, configJSON, metric.CreatedAt)
	if err != nil {
		return server.Metric{}, err
	}

	metric.ID = id
	return metric, nil
}

func (r *Repository) GetMetric(id string) (server.Metric, bool) {
	query := `
		SELECT id, name, type, config, created_at
		FROM metrics
		WHERE id = $1
	`

	var metric server.Metric
	var configJSON []byte

	err := r.db.QueryRow(query, id).Scan(&metric.ID, &metric.Name, &metric.Type, &configJSON, &metric.CreatedAt)
	if err != nil {
		return server.Metric{}, false
	}

	json.Unmarshal(configJSON, &metric.Config)
	return metric, true
}

func (r *Repository) ListMetrics() []server.Metric {
	query := `
		SELECT id, name, type, config, created_at
		FROM metrics
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return []server.Metric{}
	}
	defer rows.Close()

	items := make([]server.Metric, 0)
	for rows.Next() {
		var metric server.Metric
		var configJSON []byte

		if err := rows.Scan(&metric.ID, &metric.Name, &metric.Type, &configJSON, &metric.CreatedAt); err != nil {
			continue
		}

		json.Unmarshal(configJSON, &metric.Config)
		items = append(items, metric)
	}

	return items
}

func (r *Repository) CreateExperiment(exp server.Experiment) (server.Experiment, error) {
	id := exp.ID
	if id == "" {
		id = fmt.Sprintf("exp_%d", time.Now().UnixNano())
	}
	taskIDsJSON, _ := json.Marshal(exp.TaskIDs)
	agentIDsJSON, _ := json.Marshal(exp.AgentIDs)
	variantsJSON, _ := json.Marshal(exp.Variants)

	query := `
		INSERT INTO experiments (id, name, description, task_ids, agent_ids, variants, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, id, exp.Name, exp.Description, taskIDsJSON, agentIDsJSON, variantsJSON, exp.Status, exp.CreatedAt, exp.UpdatedAt)
	if err != nil {
		return server.Experiment{}, err
	}
	exp.ID = id
	return exp, nil
}

func (r *Repository) GetExperiment(id string) (server.Experiment, bool) {
	query := `
		SELECT id, name, description, task_ids, agent_ids, variants, status, created_at, updated_at
		FROM experiments WHERE id = $1
	`
	var exp server.Experiment
	var taskIDsJSON, agentIDsJSON, variantsJSON []byte
	err := r.db.QueryRow(query, id).Scan(&exp.ID, &exp.Name, &exp.Description, &taskIDsJSON, &agentIDsJSON, &variantsJSON, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt)
	if err != nil {
		return server.Experiment{}, false
	}
	json.Unmarshal(taskIDsJSON, &exp.TaskIDs)
	json.Unmarshal(agentIDsJSON, &exp.AgentIDs)
	json.Unmarshal(variantsJSON, &exp.Variants)
	return exp, true
}

func (r *Repository) ListExperiments() []server.Experiment {
	query := `SELECT id, name, description, task_ids, agent_ids, variants, status, created_at, updated_at FROM experiments ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return []server.Experiment{}
	}
	defer rows.Close()

	items := make([]server.Experiment, 0)
	for rows.Next() {
		var exp server.Experiment
		var taskIDsJSON, agentIDsJSON, variantsJSON []byte
		if err := rows.Scan(&exp.ID, &exp.Name, &exp.Description, &taskIDsJSON, &agentIDsJSON, &variantsJSON, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal(taskIDsJSON, &exp.TaskIDs)
		json.Unmarshal(agentIDsJSON, &exp.AgentIDs)
		json.Unmarshal(variantsJSON, &exp.Variants)
		items = append(items, exp)
	}
	return items
}

func (r *Repository) CreateExperimentRun(run server.ExperimentRun) (server.ExperimentRun, error) {
	id := run.ID
	if id == "" {
		id = fmt.Sprintf("run_%d", time.Now().UnixNano())
	}
	metricScoresJSON, _ := json.Marshal(run.MetricScores)

	query := `
		INSERT INTO experiment_runs (id, experiment_id, task_id, agent_id, variant_id, metric_scores, overall_score, duration_ms, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(query, id, run.ExperimentID, run.TaskID, run.AgentID, run.VariantID, metricScoresJSON, run.OverallScore, run.DurationMs, run.Status, run.CreatedAt)
	if err != nil {
		return server.ExperimentRun{}, err
	}
	run.ID = id
	return run, nil
}

func (r *Repository) ListExperimentRuns(experimentID string) []server.ExperimentRun {
	query := `
		SELECT id, experiment_id, task_id, agent_id, variant_id, metric_scores, overall_score, duration_ms, status, created_at
		FROM experiment_runs
		WHERE ($1 = '' OR experiment_id = $1)
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, experimentID)
	if err != nil {
		return []server.ExperimentRun{}
	}
	defer rows.Close()

	items := make([]server.ExperimentRun, 0)
	for rows.Next() {
		var run server.ExperimentRun
		var metricScoresJSON []byte
		if err := rows.Scan(&run.ID, &run.ExperimentID, &run.TaskID, &run.AgentID, &run.VariantID, &metricScoresJSON, &run.OverallScore, &run.DurationMs, &run.Status, &run.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal(metricScoresJSON, &run.MetricScores)
		items = append(items, run)
	}
	return items
}

func (r *Repository) UpdateExperimentRun(run server.ExperimentRun) error {
	metricScoresJSON, _ := json.Marshal(run.MetricScores)
	query := `
		UPDATE experiment_runs
		SET variant_id = $2, metric_scores = $3, overall_score = $4, duration_ms = $5, status = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(query, run.ID, run.VariantID, metricScoresJSON, run.OverallScore, run.DurationMs, run.Status)
	return err
}

func (r *Repository) CreateReplaySnapshot(snap server.ReplaySnapshot) (server.ReplaySnapshot, error) {
	id := snap.ID
	if id == "" {
		id = fmt.Sprintf("rsnap_%d", time.Now().UnixNano())
	}
	stateJSON, _ := json.Marshal(snap.StateSnapshot)
	inputSeedJSON, _ := json.Marshal(snap.InputSeed)

	query := `
		INSERT INTO replay_snapshots (id, execution_id, skill_id, skill_version, step_index, state_snapshot, input_seed, deterministic_seed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(query, id, snap.ExecutionID, snap.SkillID, snap.SkillVersion, snap.StepIndex, stateJSON, inputSeedJSON, snap.DeterministicSeed, snap.CreatedAt)
	if err != nil {
		return server.ReplaySnapshot{}, err
	}
	snap.ID = id
	return snap, nil
}

func (r *Repository) GetReplaySnapshot(id string) (server.ReplaySnapshot, bool) {
	query := `
		SELECT id, execution_id, skill_id, skill_version, step_index, state_snapshot, input_seed, deterministic_seed, created_at
		FROM replay_snapshots WHERE id = $1
	`
	var snap server.ReplaySnapshot
	var stateJSON, inputSeedJSON []byte
	err := r.db.QueryRow(query, id).Scan(&snap.ID, &snap.ExecutionID, &snap.SkillID, &snap.SkillVersion, &snap.StepIndex, &stateJSON, &inputSeedJSON, &snap.DeterministicSeed, &snap.CreatedAt)
	if err != nil {
		return server.ReplaySnapshot{}, false
	}
	json.Unmarshal(stateJSON, &snap.StateSnapshot)
	json.Unmarshal(inputSeedJSON, &snap.InputSeed)
	return snap, true
}

func (r *Repository) ListReplaySnapshots(executionID string) []server.ReplaySnapshot {
	query := `
		SELECT id, execution_id, skill_id, skill_version, step_index, state_snapshot, input_seed, deterministic_seed, created_at
		FROM replay_snapshots
		WHERE ($1 = '' OR execution_id = $1)
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, executionID)
	if err != nil {
		return []server.ReplaySnapshot{}
	}
	defer rows.Close()

	items := make([]server.ReplaySnapshot, 0)
	for rows.Next() {
		var snap server.ReplaySnapshot
		var stateJSON, inputSeedJSON []byte
		if err := rows.Scan(&snap.ID, &snap.ExecutionID, &snap.SkillID, &snap.SkillVersion, &snap.StepIndex, &stateJSON, &inputSeedJSON, &snap.DeterministicSeed, &snap.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal(stateJSON, &snap.StateSnapshot)
		json.Unmarshal(inputSeedJSON, &snap.InputSeed)
		items = append(items, snap)
	}
	return items
}

func (r *Repository) CreateTaskExecution(ex server.TaskExecution) (server.TaskExecution, error) {
	id := ex.ID
	if id == "" {
		id = fmt.Sprintf("texec_%d", time.Now().UnixNano())
	}
	inputJSON, _ := json.Marshal(ex.Input)
	outputJSON, _ := json.Marshal(ex.Output)

	query := `
		INSERT INTO task_executions (id, task_id, agent_id, status, input, output, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, id, ex.TaskID, ex.AgentID, ex.Status, inputJSON, outputJSON, ex.DurationMs, ex.CreatedAt)
	if err != nil {
		return server.TaskExecution{}, err
	}
	ex.ID = id
	return ex, nil
}

func (r *Repository) GetTaskExecution(id string) (server.TaskExecution, bool) {
	query := `
		SELECT id, task_id, agent_id, status, input, output, duration_ms, created_at
		FROM task_executions WHERE id = $1
	`
	var ex server.TaskExecution
	var inputJSON, outputJSON []byte
	err := r.db.QueryRow(query, id).Scan(&ex.ID, &ex.TaskID, &ex.AgentID, &ex.Status, &inputJSON, &outputJSON, &ex.DurationMs, &ex.CreatedAt)
	if err != nil {
		return server.TaskExecution{}, false
	}
	json.Unmarshal(inputJSON, &ex.Input)
	json.Unmarshal(outputJSON, &ex.Output)
	return ex, true
}

func (r *Repository) ListTaskExecutions() []server.TaskExecution {
	query := `SELECT id, task_id, agent_id, status, input, output, duration_ms, created_at FROM task_executions ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return []server.TaskExecution{}
	}
	defer rows.Close()

	items := make([]server.TaskExecution, 0)
	for rows.Next() {
		var ex server.TaskExecution
		var inputJSON, outputJSON []byte
		if err := rows.Scan(&ex.ID, &ex.TaskID, &ex.AgentID, &ex.Status, &inputJSON, &outputJSON, &ex.DurationMs, &ex.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal(inputJSON, &ex.Input)
		json.Unmarshal(outputJSON, &ex.Output)
		items = append(items, ex)
	}
	return items
}

func (r *Repository) UpdateTaskExecution(ex server.TaskExecution) error {
	outputJSON, _ := json.Marshal(ex.Output)
	query := `UPDATE task_executions SET status = $2, output = $3, duration_ms = $4 WHERE id = $1`
	_, err := r.db.Exec(query, ex.ID, ex.Status, outputJSON, ex.DurationMs)
	return err
}

func (r *Repository) CreateEvaluationResult(eval server.Evaluation) (server.Evaluation, error) {
	id := eval.ID
	if id == "" {
		id = fmt.Sprintf("eval_%d", time.Now().UnixNano())
	}
	detailsJSON, _ := json.Marshal(eval.Details)

	query := `
		INSERT INTO evaluations (id, task_execution_id, metric_id, score, details, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, id, eval.TaskExecutionID, eval.MetricID, eval.Score, detailsJSON, eval.EvaluatedAt)
	if err != nil {
		return server.Evaluation{}, err
	}
	eval.ID = id
	return eval, nil
}

func (r *Repository) ListEvaluations(taskExecutionID string) []server.Evaluation {
	query := `
		SELECT id, task_execution_id, metric_id, score, details, evaluated_at
		FROM evaluations
		WHERE ($1 = '' OR task_execution_id = $1)
		ORDER BY evaluated_at DESC
	`
	rows, err := r.db.Query(query, taskExecutionID)
	if err != nil {
		return []server.Evaluation{}
	}
	defer rows.Close()

	items := make([]server.Evaluation, 0)
	for rows.Next() {
		var eval server.Evaluation
		var detailsJSON []byte
		if err := rows.Scan(&eval.ID, &eval.TaskExecutionID, &eval.MetricID, &eval.Score, &detailsJSON, &eval.EvaluatedAt); err != nil {
			continue
		}
		json.Unmarshal(detailsJSON, &eval.Details)
		items = append(items, eval)
	}
	return items
}

func (r *Repository) GetUserByUsername(username string) (models.User, bool) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var user models.User
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, false
	}
	return user, true
}

func (r *Repository) GetUserByID(id string) (models.User, bool) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, false
	}
	return user, true
}

func (r *Repository) CreateUser(user models.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *Repository) CreateGatewayAPIKey(key server.GatewayAPIKey) error {
	query := `
		INSERT INTO gateway_api_keys (id, user_id, name, key_prefix, key_hash, created_at, last_used_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, key.ID, key.UserID, key.Name, key.KeyPrefix, key.KeyHash, key.CreatedAt, key.LastUsedAt, key.RevokedAt)
	return err
}

func (r *Repository) ListGatewayAPIKeys(userID string) []server.GatewayAPIKey {
	query := `
		SELECT k.id, k.user_id, k.name, k.key_prefix, k.created_at, k.last_used_at, k.revoked_at
		FROM gateway_api_keys k
		WHERE ($1 = '' OR k.user_id = $1)
		ORDER BY k.created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return []server.GatewayAPIKey{}
	}
	defer rows.Close()

	items := make([]server.GatewayAPIKey, 0)
	for rows.Next() {
		var item server.GatewayAPIKey
		var lastUsedAt sql.NullTime
		var revokedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.KeyPrefix, &item.CreatedAt, &lastUsedAt, &revokedAt); err != nil {
			continue
		}
		if lastUsedAt.Valid {
			t := lastUsedAt.Time
			item.LastUsedAt = &t
		}
		if revokedAt.Valid {
			t := revokedAt.Time
			item.RevokedAt = &t
		}
		items = append(items, item)
	}
	return items
}

func (r *Repository) GetGatewayAPIKeyByHash(keyHash string) (server.GatewayAPIKey, bool) {
	query := `
		SELECT
			k.id,
			k.user_id,
			k.name,
			k.key_prefix,
			k.key_hash,
			k.created_at,
			k.last_used_at,
			k.revoked_at,
			COALESCE(u.username, ''),
			COALESCE(u.role, '')
		FROM gateway_api_keys k
		LEFT JOIN users u ON u.id = k.user_id
		WHERE k.key_hash = $1
		  AND k.revoked_at IS NULL
	`
	var item server.GatewayAPIKey
	var lastUsedAt sql.NullTime
	var revokedAt sql.NullTime
	if err := r.db.QueryRow(query, keyHash).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.KeyPrefix,
		&item.KeyHash,
		&item.CreatedAt,
		&lastUsedAt,
		&revokedAt,
		&item.Username,
		&item.Role,
	); err != nil {
		return server.GatewayAPIKey{}, false
	}
	if lastUsedAt.Valid {
		t := lastUsedAt.Time
		item.LastUsedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		item.RevokedAt = &t
	}
	return item, true
}

func (r *Repository) TouchGatewayAPIKeyLastUsed(id string, usedAt time.Time) error {
	query := `UPDATE gateway_api_keys SET last_used_at = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, usedAt)
	return err
}

func (r *Repository) RevokeGatewayAPIKey(id string, userID string) error {
	query := `
		UPDATE gateway_api_keys
		SET revoked_at = NOW()
		WHERE id = $1
		  AND ($2 = '' OR user_id = $2)
	`
	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *Repository) CreateGatewayUsageEvent(event server.GatewayUsageEvent) error {
	query := `
		INSERT INTO gateway_usage_events (
			id,
			api_key_id,
			user_id,
			provider,
			model,
			endpoint,
			prompt_tokens,
			completion_tokens,
			total_tokens,
			cost,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(
		query,
		event.ID,
		event.APIKeyID,
		event.UserID,
		event.Provider,
		event.Model,
		event.Endpoint,
		event.PromptTokens,
		event.CompletionTokens,
		event.TotalTokens,
		event.Cost,
		event.CreatedAt,
	)
	return err
}

func (r *Repository) ListGatewayUsageByUser(userID string, from, to *time.Time) []server.GatewayUsageAggregate {
	query := `
		SELECT
			e.user_id,
			COALESCE(u.username, ''),
			e.provider,
			e.model,
			COUNT(*) AS request_count,
			COALESCE(SUM(e.prompt_tokens), 0) AS prompt_tokens,
			COALESCE(SUM(e.completion_tokens), 0) AS completion_tokens,
			COALESCE(SUM(e.total_tokens), 0) AS total_tokens,
			COALESCE(SUM(e.cost), 0) AS total_cost
		FROM gateway_usage_events e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.user_id = $1
		  AND ($2::timestamptz IS NULL OR e.created_at >= $2::timestamptz)
		  AND ($3::timestamptz IS NULL OR e.created_at <= $3::timestamptz)
		GROUP BY e.user_id, u.username, e.provider, e.model
		ORDER BY total_tokens DESC
	`
	rows, err := r.db.Query(query, userID, from, to)
	if err != nil {
		return []server.GatewayUsageAggregate{}
	}
	defer rows.Close()

	items := make([]server.GatewayUsageAggregate, 0)
	for rows.Next() {
		var item server.GatewayUsageAggregate
		if err := rows.Scan(
			&item.UserID,
			&item.Username,
			&item.Provider,
			&item.Model,
			&item.RequestCount,
			&item.PromptTokens,
			&item.CompletionTokens,
			&item.TotalTokens,
			&item.TotalCost,
		); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *Repository) ListGatewayUsageByUsers(from, to *time.Time) []server.GatewayUserUsageAggregate {
	query := `
		SELECT
			e.user_id,
			COALESCE(u.username, '') AS username,
			COUNT(*) AS request_count,
			COALESCE(SUM(e.prompt_tokens), 0) AS prompt_tokens,
			COALESCE(SUM(e.completion_tokens), 0) AS completion_tokens,
			COALESCE(SUM(e.total_tokens), 0) AS total_tokens,
			COALESCE(SUM(e.cost), 0) AS total_cost
		FROM gateway_usage_events e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE ($1::timestamptz IS NULL OR e.created_at >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR e.created_at <= $2::timestamptz)
		GROUP BY e.user_id, u.username
		ORDER BY total_tokens DESC
	`
	rows, err := r.db.Query(query, from, to)
	if err != nil {
		return []server.GatewayUserUsageAggregate{}
	}
	defer rows.Close()

	items := make([]server.GatewayUserUsageAggregate, 0)
	for rows.Next() {
		var item server.GatewayUserUsageAggregate
		if err := rows.Scan(
			&item.UserID,
			&item.Username,
			&item.RequestCount,
			&item.PromptTokens,
			&item.CompletionTokens,
			&item.TotalTokens,
			&item.TotalCost,
		); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *Repository) GetRandomTip() (server.Tip, bool) {
	query := `
		SELECT id, content, author, category, is_active, created_at, updated_at
		FROM tips
		WHERE is_active = TRUE
		ORDER BY RANDOM()
		LIMIT 1
	`
	var tip server.Tip
	var author sql.NullString
	if err := r.db.QueryRow(query).Scan(
		&tip.ID,
		&tip.Content,
		&author,
		&tip.Category,
		&tip.IsActive,
		&tip.CreatedAt,
		&tip.UpdatedAt,
	); err != nil {
		return server.Tip{}, false
	}
	if author.Valid {
		tip.Author = author.String
	}
	return tip, true
}

func (r *Repository) ListMCPServers() []server.MCPServer {
	query := `
		SELECT id, name, description, transport_type, command, args, env, url, headers,
		       status, rejection_reason, approved_by, approved_at, is_enabled,
		       created_at, updated_at
		FROM mcp_servers
		WHERE status IN ('pending', 'approved', 'disabled')
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return []server.MCPServer{}
	}
	defer rows.Close()

	return scanMCPServers(rows)
}

func (r *Repository) GetMCPServer(id string) (server.MCPServer, bool) {
	query := `
		SELECT id, name, description, transport_type, command, args, env, url, headers,
		       status, rejection_reason, approved_by, approved_at, is_enabled,
		       created_at, updated_at
		FROM mcp_servers
		WHERE id = $1
	`
	var s server.MCPServer
	var desc, cmd, url, rejectionReason, approvedBy sql.NullString
	var args, env, headers []byte
	var approvedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&s.ID,
		&s.Name,
		&desc,
		&s.TransportType,
		&cmd,
		&args,
		&env,
		&url,
		&headers,
		&s.Status,
		&rejectionReason,
		&approvedBy,
		&approvedAt,
		&s.IsEnabled,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return server.MCPServer{}, false
	}

	if desc.Valid {
		s.Description = desc.String
	}
	if cmd.Valid {
		s.Command = cmd.String
	}
	if url.Valid {
		s.URL = url.String
	}
	if rejectionReason.Valid {
		s.RejectionReason = rejectionReason.String
	}
	if approvedBy.Valid {
		s.ApprovedBy = approvedBy.String
	}
	if approvedAt.Valid {
		s.ApprovedAt = &approvedAt.Time
	}
	if args != nil {
		_ = json.Unmarshal(args, &s.Args)
	}
	if env != nil {
		_ = json.Unmarshal(env, &s.Env)
	}
	if headers != nil {
		_ = json.Unmarshal(headers, &s.Headers)
	}

	return s, true
}

func (r *Repository) CreateMCPServer(s server.MCPServer) (server.MCPServer, error) {
	id := s.ID
	if id == "" {
		id = fmt.Sprintf("mcp_%d", time.Now().UnixNano())
	}

	argsJSON, _ := json.Marshal(s.Args)
	envJSON, _ := json.Marshal(s.Env)
	headersJSON, _ := json.Marshal(s.Headers)

	query := `
		INSERT INTO mcp_servers (id, name, description, transport_type, command, args, env, url, headers, status, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		id,
		s.Name,
		s.Description,
		s.TransportType,
		s.Command,
		argsJSON,
		envJSON,
		s.URL,
		headersJSON,
		s.Status,
		s.IsEnabled,
	).Scan(&s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return server.MCPServer{}, fmt.Errorf("create mcp server: %w", err)
	}

	s.ID = id
	return s, nil
}

func (r *Repository) UpdateMCPServer(s server.MCPServer) error {
	argsJSON, _ := json.Marshal(s.Args)
	envJSON, _ := json.Marshal(s.Env)
	headersJSON, _ := json.Marshal(s.Headers)

	query := `
		UPDATE mcp_servers
		SET name = $2, description = $3, transport_type = $4, command = $5,
		    args = $6, env = $7, url = $8, headers = $9, status = $10,
		    is_enabled = $11, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(
		query,
		s.ID,
		s.Name,
		s.Description,
		s.TransportType,
		s.Command,
		argsJSON,
		envJSON,
		s.URL,
		headersJSON,
		s.Status,
		s.IsEnabled,
	)
	return err
}

func (r *Repository) DeleteMCPServer(id string) error {
	_, err := r.db.Exec("DELETE FROM mcp_servers WHERE id = $1", id)
	return err
}

func scanMCPServers(rows *sql.Rows) []server.MCPServer {
	servers := make([]server.MCPServer, 0)
	for rows.Next() {
		var s server.MCPServer
		var desc, cmd, url, rejectionReason, approvedBy sql.NullString
		var args, env, headers []byte
		var approvedAt sql.NullTime

		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&desc,
			&s.TransportType,
			&cmd,
			&args,
			&env,
			&url,
			&headers,
			&s.Status,
			&rejectionReason,
			&approvedBy,
			&approvedAt,
			&s.IsEnabled,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			continue
		}

		if desc.Valid {
			s.Description = desc.String
		}
		if cmd.Valid {
			s.Command = cmd.String
		}
		if url.Valid {
			s.URL = url.String
		}
		if rejectionReason.Valid {
			s.RejectionReason = rejectionReason.String
		}
		if approvedBy.Valid {
			s.ApprovedBy = approvedBy.String
		}
		if approvedAt.Valid {
			s.ApprovedAt = &approvedAt.Time
		}
		if args != nil {
			_ = json.Unmarshal(args, &s.Args)
		}
		if env != nil {
			_ = json.Unmarshal(env, &s.Env)
		}
		if headers != nil {
			_ = json.Unmarshal(headers, &s.Headers)
		}

		servers = append(servers, s)
	}
	return servers
}

func (r *Repository) CreateGatewayProviderKey(key server.GatewayProviderKey) error {
	query := `
		INSERT INTO gateway_provider_keys (id, user_id, provider, model, key_hash, key_prefix, encrypted_key, is_enabled, monthly_limit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(query, key.ID, key.UserID, key.Provider, key.Model, key.KeyHash, key.KeyPrefix, key.EncryptedKey, key.IsEnabled, key.MonthlyLimit, key.CreatedAt, key.UpdatedAt)
	return err
}

func (r *Repository) ListGatewayProviderKeys(userID string) []server.GatewayProviderKey {
	query := `
		SELECT id, user_id, provider, model, key_prefix, is_enabled, monthly_limit, created_at, updated_at
		FROM gateway_provider_keys
		WHERE $1 = '' OR user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return []server.GatewayProviderKey{}
	}
	defer rows.Close()

	items := make([]server.GatewayProviderKey, 0)
	for rows.Next() {
		var item server.GatewayProviderKey
		var model sql.NullString
		if err := rows.Scan(&item.ID, &item.UserID, &item.Provider, &model, &item.KeyPrefix, &item.IsEnabled, &item.MonthlyLimit, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		if model.Valid {
			item.Model = model.String
		}
		items = append(items, item)
	}
	return items
}

func (r *Repository) GetGatewayProviderKey(id string) (server.GatewayProviderKey, bool) {
	query := `
		SELECT id, user_id, provider, model, key_prefix, is_enabled, monthly_limit, created_at, updated_at
		FROM gateway_provider_keys
		WHERE id = $1
	`
	var item server.GatewayProviderKey
	var model sql.NullString
	if err := r.db.QueryRow(query, id).Scan(&item.ID, &item.UserID, &item.Provider, &model, &item.KeyPrefix, &item.IsEnabled, &item.MonthlyLimit, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return server.GatewayProviderKey{}, false
	}
	if model.Valid {
		item.Model = model.String
	}
	return item, true
}

func (r *Repository) GetGatewayProviderKeyByHash(keyHash string) (server.GatewayProviderKey, bool) {
	query := `
		SELECT id, user_id, provider, model, key_prefix, is_enabled, monthly_limit, created_at, updated_at
		FROM gateway_provider_keys
		WHERE key_hash = $1
	`
	var item server.GatewayProviderKey
	var model sql.NullString
	if err := r.db.QueryRow(query, keyHash).Scan(&item.ID, &item.UserID, &item.Provider, &model, &item.KeyPrefix, &item.IsEnabled, &item.MonthlyLimit, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return server.GatewayProviderKey{}, false
	}
	if model.Valid {
		item.Model = model.String
	}
	return item, true
}

func (r *Repository) UpdateGatewayProviderKey(key server.GatewayProviderKey) error {
	query := `
		UPDATE gateway_provider_keys
		SET provider = $2, model = $3, key_hash = $4, key_prefix = $5, encrypted_key = $6, is_enabled = $7, monthly_limit = $8, updated_at = $9
		WHERE id = $1
	`
	_, err := r.db.Exec(query, key.ID, key.Provider, key.Model, key.KeyHash, key.KeyPrefix, key.EncryptedKey, key.IsEnabled, key.MonthlyLimit, key.UpdatedAt)
	return err
}

func (r *Repository) DeleteGatewayProviderKey(id, userID string) error {
	query := `DELETE FROM gateway_provider_keys WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *Repository) ToggleGatewayProviderKey(id, userID string, enabled bool) error {
	query := `UPDATE gateway_provider_keys SET is_enabled = $3, updated_at = NOW() WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, id, userID, enabled)
	return err
}

func (r *Repository) GetUserBudget(userID string) (server.UserBudget, bool) {
	query := `
		SELECT id, user_id, monthly_limit, daily_limit, used_this_month, used_today, reset_at, updated_at
		FROM user_budgets
		WHERE user_id = $1
	`
	var item server.UserBudget
	if err := r.db.QueryRow(query, userID).Scan(&item.ID, &item.UserID, &item.MonthlyLimit, &item.DailyLimit, &item.UsedThisMonth, &item.UsedToday, &item.ResetAt, &item.UpdatedAt); err != nil {
		return server.UserBudget{}, false
	}
	return item, true
}

func (r *Repository) CreateOrUpdateUserBudget(budget server.UserBudget) error {
	query := `
		INSERT INTO user_budgets (id, user_id, monthly_limit, daily_limit, used_this_month, used_today, reset_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			monthly_limit = EXCLUDED.monthly_limit,
			daily_limit = EXCLUDED.daily_limit,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, budget.ID, budget.UserID, budget.MonthlyLimit, budget.DailyLimit, budget.UsedThisMonth, budget.UsedToday, budget.ResetAt, budget.UpdatedAt)
	return err
}

func (r *Repository) DecrementUserBudget(userID string, tokens int) (int, error) {
	query := `UPDATE user_budgets SET used_this_month = used_this_month + $2, used_today = used_today + $2, updated_at = NOW() WHERE user_id = $1`
	_, err := r.db.Exec(query, userID, tokens)
	return 0, err
}
