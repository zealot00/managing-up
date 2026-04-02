CREATE TABLE IF NOT EXISTS seh_datasets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    owner TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    manifest JSONB NOT NULL DEFAULT '{}',
    case_count INTEGER NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seh_cases (
    id TEXT PRIMARY KEY,
    dataset_id TEXT NOT NULL REFERENCES seh_datasets(id) ON DELETE CASCADE,
    skill TEXT NOT NULL,
    source TEXT NOT NULL,
    status TEXT NOT NULL,
    provenance JSONB NOT NULL DEFAULT '{}',
    input JSONB NOT NULL DEFAULT '{}',
    expected JSONB NOT NULL DEFAULT '{}',
    tags JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seh_runs (
    id TEXT PRIMARY KEY,
    dataset_id TEXT NOT NULL REFERENCES seh_datasets(id),
    skill TEXT NOT NULL,
    runtime TIMESTAMPTZ NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}',
    results JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seh_policies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    require_provenance BOOLEAN NOT NULL DEFAULT FALSE,
    require_approved_for_score BOOLEAN NOT NULL DEFAULT FALSE,
    min_source_diversity INTEGER NOT NULL DEFAULT 1,
    min_golden_weight DOUBLE PRECISION NOT NULL DEFAULT 0,
    source_policies JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seh_releases (
    id TEXT PRIMARY KEY,
    skill_id TEXT NOT NULL,
    version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_approval',
    artifacts JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_seh_cases_dataset_id ON seh_cases(dataset_id);
CREATE INDEX idx_seh_cases_skill ON seh_cases(skill);
CREATE INDEX idx_seh_runs_dataset_id ON seh_runs(dataset_id);
CREATE INDEX idx_seh_runs_created_at ON seh_runs(created_at);
