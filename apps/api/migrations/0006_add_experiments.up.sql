-- experiments: Collection of experiment runs for comparing agent/skill performance
CREATE TABLE IF NOT EXISTS experiments (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    task_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    agent_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_experiments_status ON experiments(status);

-- experiment_runs: Individual run within an experiment
CREATE TABLE IF NOT EXISTS experiment_runs (
    id TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    metric_scores JSONB NOT NULL DEFAULT '{}'::jsonb,
    overall_score REAL NOT NULL DEFAULT 0,
    duration_ms BIGINT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_experiment_runs_experiment_id ON experiment_runs(experiment_id);
CREATE INDEX IF NOT EXISTS idx_experiment_runs_task_id ON experiment_runs(task_id);
CREATE INDEX IF NOT EXISTS idx_experiment_runs_agent_id ON experiment_runs(agent_id);

-- replay_snapshots: Execution state snapshots for deterministic replay
CREATE TABLE IF NOT EXISTS replay_snapshots (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    skill_id TEXT NOT NULL,
    skill_version TEXT NOT NULL,
    step_index INT NOT NULL,
    state_snapshot JSONB NOT NULL,
    input_seed JSONB NOT NULL,
    deterministic_seed BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_replay_snapshots_execution_id ON replay_snapshots(execution_id);
CREATE INDEX IF NOT EXISTS idx_replay_snapshots_deterministic_seed ON replay_snapshots(deterministic_seed);
