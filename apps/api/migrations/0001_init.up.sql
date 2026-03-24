CREATE TABLE IF NOT EXISTS skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    owner_team TEXT NOT NULL,
    risk_level TEXT NOT NULL,
    status TEXT NOT NULL,
    current_version TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS skill_versions (
    id TEXT PRIMARY KEY,
    skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    status TEXT NOT NULL,
    change_summary TEXT NOT NULL,
    approval_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS procedure_drafts (
    id TEXT PRIMARY KEY,
    procedure_key TEXT NOT NULL,
    title TEXT NOT NULL,
    validation_status TEXT NOT NULL,
    required_tools JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS executions (
    id TEXT PRIMARY KEY,
    skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    skill_name TEXT NOT NULL,
    status TEXT NOT NULL,
    triggered_by TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    current_step_id TEXT NOT NULL,
    input JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS approvals (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    skill_name TEXT NOT NULL,
    step_id TEXT NOT NULL,
    status TEXT NOT NULL,
    approver_group TEXT NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skills_status ON skills(status);
CREATE INDEX IF NOT EXISTS idx_skill_versions_skill_id ON skill_versions(skill_id);
CREATE INDEX IF NOT EXISTS idx_procedure_drafts_status ON procedure_drafts(validation_status);
CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status);

