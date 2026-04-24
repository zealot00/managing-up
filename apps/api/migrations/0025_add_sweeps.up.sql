-- Sweep Engine: Hyperparameter sweep configuration and execution tracking
CREATE TABLE sweep_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    task_id         VARCHAR(255) NOT NULL,
    parameters_     JSONB NOT NULL DEFAULT '{}',
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_runs      INTEGER NOT NULL DEFAULT 0,
    completed       INTEGER NOT NULL DEFAULT 0,
    created_by      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sweep_configs_task ON sweep_configs(task_id);
CREATE INDEX idx_sweep_configs_status ON sweep_configs(status);
CREATE INDEX idx_sweep_configs_created_by ON sweep_configs(created_by);

CREATE TABLE sweep_runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sweep_config_id  UUID NOT NULL REFERENCES sweep_configs(id) ON DELETE CASCADE,
    variant_index     INTEGER NOT NULL,
    model             VARCHAR(100) NOT NULL,
    temperature      FLOAT NOT NULL,
    max_tokens       INTEGER NOT NULL,
    prompt_id        VARCHAR(100) NOT NULL,
    prompt_label     TEXT NOT NULL,
    status           VARCHAR(50) NOT NULL DEFAULT 'pending',
    task_execution_id VARCHAR(255),
    score            FLOAT,
    duration_ms      BIGINT,
    error            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at     TIMESTAMPTZ
);

CREATE INDEX idx_sweep_runs_config ON sweep_runs(sweep_config_id);
CREATE INDEX idx_sweep_runs_status ON sweep_runs(status);
CREATE INDEX idx_sweep_runs_execution ON sweep_runs(task_execution_id);
