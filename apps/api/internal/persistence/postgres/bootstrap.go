package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Migrate applies all up migration files in lexical order.
func Migrate(dsn, migrationsDir string) error {
	db, err := openDB(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}

	sort.Strings(files)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(file), err)
		}
	}

	return nil
}

// Seed inserts deterministic bootstrap data for local and test environments.
func Seed(dsn string) error {
	db, err := openDB(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()

	if err := seedSkills(tx, now); err != nil {
		return err
	}
	if err := seedSkillVersions(tx, now); err != nil {
		return err
	}
	if err := seedProcedureDrafts(tx, now); err != nil {
		return err
	}
	if err := seedExecutions(tx, now); err != nil {
		return err
	}
	if err := seedApprovals(tx, now); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}

func openDB(dsn string) (*sql.DB, error) {
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

	return db, nil
}

func seedSkills(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO skills (id, name, owner_team, risk_level, status, current_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			owner_team = EXCLUDED.owner_team,
			risk_level = EXCLUDED.risk_level,
			status = EXCLUDED.status,
			current_version = EXCLUDED.current_version,
			updated_at = EXCLUDED.updated_at
	`

	records := [][]any{
		{"skill_001", "restart_service_skill", "platform_team", "medium", "published", "v1", now.Add(-72 * time.Hour), now},
		{"skill_002", "collect_logs_skill", "sre_team", "low", "published", "v3", now.Add(-48 * time.Hour), now},
		{"skill_003", "rollback_deployment_skill", "platform_team", "high", "draft", "", now.Add(-12 * time.Hour), now},
	}

	for _, record := range records {
		if _, err := tx.Exec(query, record...); err != nil {
			return fmt.Errorf("seed skills: %w", err)
		}
	}

	return nil
}

func seedSkillVersions(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO skill_versions (id, skill_id, version, status, change_summary, approval_required, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			skill_id = EXCLUDED.skill_id,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			change_summary = EXCLUDED.change_summary,
			approval_required = EXCLUDED.approval_required
	`

	records := [][]any{
		{"version_001", "skill_001", "v1", "published", "Initial restart automation flow.", true, now.Add(-72 * time.Hour)},
		{"version_002", "skill_002", "v3", "published", "Added export safety checks and retry handling.", true, now.Add(-48 * time.Hour)},
		{"version_003", "skill_003", "v0-draft", "draft", "Rollback flow under review.", true, now.Add(-12 * time.Hour)},
	}

	for _, record := range records {
		if _, err := tx.Exec(query, record...); err != nil {
			return fmt.Errorf("seed skill_versions: %w", err)
		}
	}

	return nil
}

func seedProcedureDrafts(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO procedure_drafts (id, procedure_key, title, validation_status, required_tools, source_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			procedure_key = EXCLUDED.procedure_key,
			title = EXCLUDED.title,
			validation_status = EXCLUDED.validation_status,
			required_tools = EXCLUDED.required_tools,
			source_type = EXCLUDED.source_type
	`

	tools1, _ := json.Marshal([]string{"monitor_api", "orchestrator_api"})
	tools2, _ := json.Marshal([]string{"shell_adapter", "storage_exporter"})
	records := [][]any{
		{"draft_001", "runbook_restart_service", "Restart Service Runbook", "validated", string(tools1), "markdown", now.Add(-96 * time.Hour)},
		{"draft_002", "collect_production_logs", "Collect Production Logs", "draft", string(tools2), "pdf", now.Add(-18 * time.Hour)},
	}

	for _, record := range records {
		if _, err := tx.Exec(query, record...); err != nil {
			return fmt.Errorf("seed procedure_drafts: %w", err)
		}
	}

	return nil
}

func seedExecutions(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO executions (id, skill_id, skill_name, status, triggered_by, started_at, current_step_id, input)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			skill_id = EXCLUDED.skill_id,
			skill_name = EXCLUDED.skill_name,
			status = EXCLUDED.status,
			triggered_by = EXCLUDED.triggered_by,
			started_at = EXCLUDED.started_at,
			current_step_id = EXCLUDED.current_step_id,
			input = EXCLUDED.input
	`

	input1, _ := json.Marshal(map[string]any{"server_id": "srv-001"})
	input2, _ := json.Marshal(map[string]any{"collection_id": "logs-20260319"})
	input3, _ := json.Marshal(map[string]any{"server_id": "srv-007"})
	records := [][]any{
		{"exec_001", "skill_001", "restart_service_skill", "running", "sre_oncall", now, "verify_health", string(input1)},
		{"exec_002", "skill_002", "collect_logs_skill", "waiting_approval", "ops_manager", now.Add(-35 * time.Minute), "approval_before_export", string(input2)},
		{"exec_003", "skill_001", "restart_service_skill", "succeeded", "platform_operator", now.Add(-2 * time.Hour), "completed", string(input3)},
	}

	for _, record := range records {
		if _, err := tx.Exec(query, record...); err != nil {
			return fmt.Errorf("seed executions: %w", err)
		}
	}

	return nil
}

func seedApprovals(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO approvals (id, execution_id, skill_name, step_id, status, approver_group, requested_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			execution_id = EXCLUDED.execution_id,
			skill_name = EXCLUDED.skill_name,
			step_id = EXCLUDED.step_id,
			status = EXCLUDED.status,
			approver_group = EXCLUDED.approver_group,
			requested_at = EXCLUDED.requested_at
	`

	record := []any{"approval_001", "exec_002", "collect_logs_skill", "approval_before_export", "waiting", "ops_manager", now.Add(-30 * time.Minute)}
	if _, err := tx.Exec(query, record...); err != nil {
		return fmt.Errorf("seed approvals: %w", err)
	}

	return nil
}

// MigrationsDir resolves the migration directory from the API module root.
func MigrationsDir(moduleRoot string) string {
	return filepath.Join(strings.TrimSpace(moduleRoot), "migrations")
}
