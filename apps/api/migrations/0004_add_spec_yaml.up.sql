-- Add spec_yaml column to skill_versions for storing skill execution spec
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS spec_yaml TEXT NOT NULL DEFAULT '';

-- Add created_by column to executions table
ALTER TABLE executions ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT '';

-- Add ended_at and duration_ms columns to executions for tracking execution time
ALTER TABLE executions ADD COLUMN IF NOT EXISTS ended_at TIMESTAMPTZ;
ALTER TABLE executions ADD COLUMN IF NOT EXISTS duration_ms BIGINT;

-- Add execution_step_id for tracking individual step executions
CREATE TABLE IF NOT EXISTS execution_steps (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_id TEXT NOT NULL,
    status TEXT NOT NULL,
    tool_ref TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    duration_ms BIGINT,
    output JSONB,
    error TEXT,
    attempt_no INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_execution_steps_execution_id ON execution_steps(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_steps_status ON execution_steps(status);
