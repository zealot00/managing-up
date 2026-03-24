-- execution_traces: Immutable append-only log of all execution events
CREATE TABLE IF NOT EXISTS execution_traces (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_id TEXT,
    event_type TEXT NOT NULL,
    event_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_traces_execution_id ON execution_traces(execution_id);
CREATE INDEX IF NOT EXISTS idx_traces_step_id ON execution_traces(step_id);
CREATE INDEX IF NOT EXISTS idx_traces_event_type ON execution_traces(event_type);
CREATE INDEX IF NOT EXISTS idx_traces_timestamp ON execution_traces(timestamp DESC);

-- tasks: Registry of reusable tasks for evaluation
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    skill_id TEXT REFERENCES skills(id) ON DELETE SET NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    difficulty TEXT NOT NULL DEFAULT 'medium',
    test_cases JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_skill_id ON tasks(skill_id);
CREATE INDEX IF NOT EXISTS idx_tasks_difficulty ON tasks(difficulty);

-- task_executions: Each run of a task by an agent
CREATE TABLE IF NOT EXISTS task_executions (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB,
    duration_ms BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_executions_task_id ON task_executions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_executions_agent_id ON task_executions(agent_id);

-- metrics: Available metric definitions
CREATE TABLE IF NOT EXISTS metrics (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);

-- evaluations: Results of running metrics against task executions
CREATE TABLE IF NOT EXISTS evaluations (
    id TEXT PRIMARY KEY,
    task_execution_id TEXT NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
    metric_id TEXT NOT NULL REFERENCES metrics(id) ON DELETE RESTRICT,
    score REAL NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_evaluations_task_execution_id ON evaluations(task_execution_id);
CREATE INDEX IF NOT EXISTS idx_evaluations_metric_id ON evaluations(metric_id);
