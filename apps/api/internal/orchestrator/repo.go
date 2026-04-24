package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(dsn string) (*Repo, error) {
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

	return &Repo{db: db}, nil
}

func (r *Repo) Close() error {
	return r.db.Close()
}

func (r *Repo) DB() *sql.DB {
	return r.db
}

func (r *Repo) CreateRun(req CreateRunRequest) (*RunDetail, error) {
	id := fmt.Sprintf("run_%d", time.Now().UnixNano())
	now := time.Now().UTC()

	sourceJSON, _ := json.Marshal(req.Source)
	optionsJSON, _ := json.Marshal(req.Options)

	query := `
		INSERT INTO orchestrator_runs (id, skill_name, source, options, status, stage, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'queued', 'extraction', $5, $6)
		RETURNING id, status, stage, skill_name, created_at, updated_at
	`

	run := &RunDetail{
		RunID:     id,
		Status:    "queued",
		Stage:     "extraction",
		SkillName: req.SkillName,
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
		Errors:    []ErrorResponse{},
	}

	_, err := r.db.Exec(query, id, req.SkillName, string(sourceJSON), string(optionsJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	return run, nil
}

func (r *Repo) GetRun(id string) (*RunDetail, error) {
	query := `
		SELECT id, status, stage, skill_name, created_at, updated_at, result, errors
		FROM orchestrator_runs
		WHERE id = $1
	`

	var run RunDetail
	var resultJSON, errorsJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&run.RunID, &run.Status, &run.Stage, &run.SkillName, &run.CreatedAt, &run.UpdatedAt, &resultJSON, &errorsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get run: %w", err)
	}

	if resultJSON != nil {
		var result RunResult
		json.Unmarshal(resultJSON, &result)
		run.Result = &result
	}

	if errorsJSON != nil {
		json.Unmarshal(errorsJSON, &run.Errors)
	}

	return &run, nil
}

func (r *Repo) UpdateRunStatus(id, status, stage string) error {
	query := `
		UPDATE orchestrator_runs
		SET status = $2, stage = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, status, stage, time.Now().UTC())
	return err
}

func (r *Repo) UpdateRunResult(id string, result *RunResult) error {
	resultJSON, _ := json.Marshal(result)
	query := `
		UPDATE orchestrator_runs
		SET result = $2, updated_at = $3
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, string(resultJSON), time.Now().UTC())
	return err
}

func (r *Repo) CreateArtifact(runID string, artifact ArtifactRef) error {
	query := `
		INSERT INTO orchestrator_artifacts (run_id, kind, uri)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, runID, artifact.Kind, artifact.URI)
	return err
}

func (r *Repo) ListRunArtifacts(runID string) ([]ArtifactRef, error) {
	query := `
		SELECT kind, uri
		FROM orchestrator_artifacts
		WHERE run_id = $1
	`

	rows, err := r.db.Query(query, runID)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []ArtifactRef
	for rows.Next() {
		var a ArtifactRef
		if err := rows.Scan(&a.Kind, &a.URI); err != nil {
			continue
		}
		artifacts = append(artifacts, a)
	}

	return artifacts, nil
}

func (r *Repo) CreateSkill(req CreateSkillRequest) (*Skill, error) {
	id := req.SkillID
	if id == "" {
		id = fmt.Sprintf("skill_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()

	tagsJSON, _ := json.Marshal(req.Tags)

	query := `
		INSERT INTO orchestrator_skills (id, name, owner, tags, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			owner = EXCLUDED.owner,
			tags = EXCLUDED.tags
		RETURNING id, name, COALESCE(owner, ''), COALESCE(tags, '[]'::jsonb), created_at
	`

	skill := &Skill{}
	var owner string
	var tagsBytes []byte

	err := r.db.QueryRow(query, id, req.Name, req.Owner, string(tagsJSON), now).Scan(
		&skill.SkillID, &skill.Name, &owner, &tagsBytes, &skill.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create skill: %w", err)
	}

	skill.Owner = owner
	json.Unmarshal(tagsBytes, &skill.Tags)

	return skill, nil
}

func (r *Repo) GetSkill(id string) (*Skill, error) {
	query := `
		SELECT id, name, COALESCE(owner, ''), COALESCE(tags, '[]'::jsonb), created_at
		FROM orchestrator_skills
		WHERE id = $1
	`

	skill := &Skill{}
	var owner string
	var tagsBytes []byte

	err := r.db.QueryRow(query, id).Scan(
		&skill.SkillID, &skill.Name, &owner, &tagsBytes, &skill.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get skill: %w", err)
	}

	skill.Owner = owner
	json.Unmarshal(tagsBytes, &skill.Tags)

	return skill, nil
}

func (r *Repo) ListSkills() ([]Skill, error) {
	query := `
		SELECT id, name, COALESCE(owner, ''), COALESCE(tags, '[]'::jsonb), created_at
		FROM orchestrator_skills
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		skill := &Skill{}
		var owner string
		var tagsBytes []byte

		if err := rows.Scan(&skill.SkillID, &skill.Name, &owner, &tagsBytes, &skill.CreatedAt); err != nil {
			continue
		}

		skill.Owner = owner
		json.Unmarshal(tagsBytes, &skill.Tags)
		skills = append(skills, *skill)
	}

	return skills, nil
}

func (r *Repo) CreateSkillVersion(skillID string, req CreateSkillVersionRequest) (*SkillVersion, error) {
	now := time.Now().UTC()

	configJSON, _ := json.Marshal(req.ConfigSnapshot)
	artifactsJSON, _ := json.Marshal(req.Artifacts)

	query := `
		INSERT INTO orchestrator_skill_versions (skill_id, version, source_hash, schema_hash, config_snapshot, artifacts, run_id, created_at, promoted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false)
		RETURNING skill_id, version, source_hash, schema_hash, config_snapshot, artifacts, COALESCE(run_id, ''), created_at, promoted
	`

	sv := &SkillVersion{}
	var configBytes, artifactsBytes []byte
	var runID string

	err := r.db.QueryRow(query, skillID, req.Version, req.SourceHash, req.SchemaHash, string(configJSON), string(artifactsJSON), req.RunID, now).Scan(
		&sv.SkillID, &sv.Version, &sv.SourceHash, &sv.SchemaHash, &configBytes, &artifactsBytes, &runID, &sv.CreatedAt, &sv.Promoted,
	)
	if err != nil {
		return nil, fmt.Errorf("create skill version: %w", err)
	}

	sv.RunID = runID
	json.Unmarshal(configBytes, &sv.ConfigSnapshot)
	json.Unmarshal(artifactsBytes, &sv.Artifacts)

	return sv, nil
}

func (r *Repo) GetSkillVersion(skillID, version string) (*SkillVersion, error) {
	query := `
		SELECT skill_id, version, source_hash, schema_hash, config_snapshot, artifacts, COALESCE(run_id, ''), created_at, promoted
		FROM orchestrator_skill_versions
		WHERE skill_id = $1 AND version = $2
	`

	sv := &SkillVersion{}
	var configBytes, artifactsBytes []byte

	err := r.db.QueryRow(query, skillID, version).Scan(
		&sv.SkillID, &sv.Version, &sv.SourceHash, &sv.SchemaHash, &configBytes, &artifactsBytes, &sv.RunID, &sv.CreatedAt, &sv.Promoted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get skill version: %w", err)
	}

	json.Unmarshal(configBytes, &sv.ConfigSnapshot)
	json.Unmarshal(artifactsBytes, &sv.Artifacts)

	return sv, nil
}

func (r *Repo) ListSkillVersions(skillID string) ([]SkillVersion, error) {
	query := `
		SELECT skill_id, version, source_hash, schema_hash, config_snapshot, artifacts, COALESCE(run_id, ''), created_at, promoted
		FROM orchestrator_skill_versions
		WHERE skill_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, skillID)
	if err != nil {
		return nil, fmt.Errorf("list skill versions: %w", err)
	}
	defer rows.Close()

	var versions []SkillVersion
	for rows.Next() {
		sv := &SkillVersion{}
		var configBytes, artifactsBytes []byte

		if err := rows.Scan(&sv.SkillID, &sv.Version, &sv.SourceHash, &sv.SchemaHash, &configBytes, &artifactsBytes, &sv.RunID, &sv.CreatedAt, &sv.Promoted); err != nil {
			continue
		}

		json.Unmarshal(configBytes, &sv.ConfigSnapshot)
		json.Unmarshal(artifactsBytes, &sv.Artifacts)
		versions = append(versions, *sv)
	}

	return versions, nil
}

func (r *Repo) UpdateSkillVersionPromoted(skillID, version string, promoted bool) error {
	query := `
		UPDATE orchestrator_skill_versions
		SET promoted = $3
		WHERE skill_id = $1 AND version = $2
	`
	_, err := r.db.Exec(query, skillID, version, promoted)
	return err
}

func (r *Repo) GetLatestPassedSnapshot(ctx context.Context, skillID, version string) (*SkillCapabilitySnapshot, error) {
	query := `
		SELECT id, skill_id, version, snapshot_type, COALESCE(dataset_id, ''), COALESCE(run_id, ''), metrics,
		       overall_score, passed, COALESCE(gate_policy_id, ''), evaluated_at, created_at
		FROM skill_capability_snapshots
		WHERE skill_id = $1 AND version = $2 AND passed = true
		ORDER BY evaluated_at DESC
		LIMIT 1
	`
	var snap SkillCapabilitySnapshot
	var metricsJSON []byte

	err := r.db.QueryRowContext(ctx, query, skillID, version).Scan(
		&snap.ID, &snap.SkillID, &snap.Version, &snap.SnapshotType,
		&snap.DatasetID, &snap.RunID, &metricsJSON,
		&snap.OverallScore, &snap.Passed, &snap.GatePolicyID,
		&snap.EvaluatedAt, &snap.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	json.Unmarshal(metricsJSON, &snap.Metrics)
	return &snap, nil
}

func (r *Repo) CreateTestRun(req CreateTestRunRequest) (*TestRun, error) {
	id := fmt.Sprintf("testrun_%d", time.Now().UnixNano())
	now := time.Now().UTC()

	query := `
		INSERT INTO orchestrator_test_runs (id, skill_id, version, runner, dataset_ref, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'queued', $6, $7)
		RETURNING id, status, created_at
	`

	tr := &TestRun{
		TestRunID: id,
		Status:    "queued",
		CreatedAt: now,
	}

	_, err := r.db.Exec(query, id, req.SkillID, req.Version, req.Runner.Type, req.DatasetRef, now, now)
	if err != nil {
		return nil, fmt.Errorf("create test run: %w", err)
	}

	return tr, nil
}

func (r *Repo) GetTestRun(id string) (*TestRun, error) {
	query := `
		SELECT id, skill_id, version, runner, COALESCE(dataset_ref, ''), status, created_at, COALESCE(updated_at, created_at), COALESCE(exit_code, 0)
		FROM orchestrator_test_runs
		WHERE id = $1
	`

	tr := &TestRun{}
	var runnerType string

	err := r.db.QueryRow(query, id).Scan(
		&tr.TestRunID, &tr.SkillID, &tr.Version, &runnerType, &tr.DatasetRef, &tr.Status, &tr.CreatedAt, &tr.UpdatedAt, &tr.ExitCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get test run: %w", err)
	}

	tr.Runner = TestRunner{Type: runnerType}

	return tr, nil
}

func (r *Repo) UpdateTestRunStatus(id, status string, exitCode int) error {
	query := `
		UPDATE orchestrator_test_runs
		SET status = $2, exit_code = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, status, exitCode, time.Now().UTC())
	return err
}

func (r *Repo) GetPolicy(id string) (*Policy, error) {
	query := `
		SELECT id, name, rules
		FROM orchestrator_policies
		WHERE id = $1
	`

	policy := &Policy{}
	var rulesJSON []byte

	err := r.db.QueryRow(query, id).Scan(&policy.PolicyID, &policy.Name, &rulesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get policy: %w", err)
	}

	json.Unmarshal(rulesJSON, &policy.Rules)

	return policy, nil
}

func (r *Repo) CreatePolicy(policy Policy) error {
	rulesJSON, _ := json.Marshal(policy.Rules)

	query := `
		INSERT INTO orchestrator_policies (id, name, rules)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			rules = EXCLUDED.rules
	`
	_, err := r.db.Exec(query, policy.PolicyID, policy.Name, string(rulesJSON))
	return err
}

func (r *Repo) InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS orchestrator_runs (
			id TEXT PRIMARY KEY,
			skill_name TEXT NOT NULL,
			source JSONB NOT NULL DEFAULT '{}'::jsonb,
			options JSONB DEFAULT '{}'::jsonb,
			status TEXT NOT NULL DEFAULT 'queued',
			stage TEXT NOT NULL DEFAULT 'extraction',
			result JSONB DEFAULT NULL,
			errors JSONB DEFAULT '[]'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS orchestrator_artifacts (
			id SERIAL PRIMARY KEY,
			run_id TEXT NOT NULL REFERENCES orchestrator_runs(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			uri TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_orchestrator_artifacts_run_id ON orchestrator_artifacts(run_id);

		CREATE TABLE IF NOT EXISTS orchestrator_skills (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			owner TEXT DEFAULT '',
			tags JSONB DEFAULT '[]'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS orchestrator_skill_versions (
			id SERIAL PRIMARY KEY,
			skill_id TEXT NOT NULL REFERENCES orchestrator_skills(id) ON DELETE CASCADE,
			version TEXT NOT NULL,
			source_hash TEXT NOT NULL,
			schema_hash TEXT NOT NULL,
			config_snapshot JSONB DEFAULT '{}'::jsonb,
			artifacts JSONB DEFAULT '[]'::jsonb,
			run_id TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			promoted BOOLEAN NOT NULL DEFAULT FALSE,
			UNIQUE(skill_id, version)
		);

		CREATE INDEX IF NOT EXISTS idx_orchestrator_skill_versions_skill_id ON orchestrator_skill_versions(skill_id);

		CREATE TABLE IF NOT EXISTS orchestrator_test_runs (
			id TEXT PRIMARY KEY,
			skill_id TEXT NOT NULL,
			version TEXT NOT NULL,
			runner TEXT NOT NULL DEFAULT '{}',
			dataset_ref TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'queued',
			exit_code INTEGER DEFAULT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS orchestrator_policies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			rules JSONB NOT NULL DEFAULT '[]'::jsonb
		);

		CREATE TABLE IF NOT EXISTS orchestrator_actions (
			id TEXT PRIMARY KEY,
			action_type TEXT NOT NULL,
			skill_id TEXT NOT NULL,
			target_version TEXT NOT NULL,
			reason TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'accepted',
			accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_orchestrator_actions_skill_id ON orchestrator_actions(skill_id);
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}

	return nil
}
