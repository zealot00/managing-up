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
	if err := seedUsers(tx, now); err != nil {
		return err
	}
	if err := seedTips(tx, now); err != nil {
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

func seedUsers(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			updated_at = EXCLUDED.updated_at
	`

	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.rsAOGe8W3p3cKo.OC"
	records := [][]any{
		{"user_admin", "admin", hash, "admin", now, now},
		{"user_operator", "operator", hash, "user", now, now},
	}

	for _, record := range records {
		if _, err := tx.Exec(query, record...); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
	}

	return nil
}

func seedTips(tx *sql.Tx, now time.Time) error {
	query := `
		INSERT INTO tips (id, content, author, category, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			author = EXCLUDED.author,
			category = EXCLUDED.category,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	type tip struct {
		id       string
		content  string
		author   string
		category string
	}

	tips := []tip{
		{"tip_001", "Talk is cheap. Show me the code.", "Linus Torvalds", "quote"},
		{"tip_002", "Any fool can write code that a computer can understand. Good programmers write code that humans can understand.", "Martin Fowler", "quote"},
		{"tip_003", "First, solve the problem. Then, write the code.", "John Johnson", "quote"},
		{"tip_004", "The best error message is the one that never shows up.", "Thomas Fuchs", "quote"},
		{"tip_005", "It works on my machine.", "", "humor"},
		{"tip_006", "There are only two hard things in CS: cache invalidation and naming things.", "Phil Karlton", "quote"},
		{"tip_007", "A good programmer looks both ways before crossing a one-way street.", "Doug Linder", "humor"},
		{"tip_008", "The most dangerous phrase in programming: We've always done it this way.", "Grace Hopper", "quote"},
		{"tip_009", "Code never lies, comments sometimes do.", "Ron Jeffries", "quote"},
		{"tip_010", "Weeks of coding can save you hours of planning.", "", "humor"},
		{"tip_011", "In theory, there's no difference between theory and practice. In practice, there is.", "Yogi Berra", "quote"},
		{"tip_012", "Programming is the art of telling another human what one wants the computer to do.", "Donald Knuth", "quote"},
		{"tip_013", "The best code is no code at all.", "Jeff Atwood", "quote"},
		{"tip_014", "Debugging is twice as hard as writing the code in the first place.", "Brian Kernighan", "quote"},
		{"tip_015", "Simplicity is the soul of efficiency.", "Wes F. Williams", "quote"},
		{"tip_016", "Make it work, make it right, make it fast.", "Kent Beck", "quote"},
		{"tip_017", "Premature optimization is the root of all evil.", "Donald Knuth", "quote"},
		{"tip_018", "Software and cathedrals are much the same — first we build them, then we pray.", "Sam Redwine", "quote"},
		{"tip_019", "The only way to go fast is to go well.", "Robert C. Martin", "quote"},
		{"tip_020", "One of my most productive days was throwing away 1000 lines of code.", "Ken Thompson", "quote"},
		{"tip_021", "Before software can be reusable it first has to be usable.", "Ralph Johnson", "quote"},
		{"tip_022", "Good design is as little design as possible.", "Dieter Rams", "quote"},
		{"tip_023", "The function of good software is to make the complex appear to be simple.", "Grady Booch", "quote"},
		{"tip_024", "Measuring programming progress by lines of code is like measuring aircraft building progress by weight.", "Bill Gates", "quote"},
		{"tip_025", "Simplicity is prerequisite for reliability.", "Edsger W. Dijkstra", "quote"},
		{"tip_026", "Deleted code is debugged code.", "Jeff Sickel", "quote"},
		{"tip_027", "If debugging is the process of removing bugs, then programming must be the process of putting them in.", "Edsger W. Dijkstra", "humor"},
		{"tip_028", "The best programmers are not marginally better than merely good ones. They are an order-of-magnitude better.", "Robert C. Martin", "quote"},
		{"tip_029", "You can't trust code that you did not totally create yourself.", "Ken Thompson", "quote"},
		{"tip_030", "Sometimes it pays to stay in bed on Monday, rather than spending the rest of the week debugging Monday's code.", "Dan Salomon", "humor"},
	}

	for i, t := range tips {
		if _, err := tx.Exec(query, t.id, t.content, t.author, t.category, true, now.Add(time.Duration(i)*time.Minute), now); err != nil {
			return fmt.Errorf("seed tips: %w", err)
		}
	}

	return nil
}

func MigrationsDir(moduleRoot string) string {
	return filepath.Join(strings.TrimSpace(moduleRoot), "migrations")
}
