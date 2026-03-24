-- Add new columns to tasks table (all nullable for backward compatibility)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS task_type TEXT NOT NULL DEFAULT 'benchmark';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS input_source TEXT NOT NULL DEFAULT 'inline';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS input_path TEXT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS input_format TEXT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS gold_type TEXT NOT NULL DEFAULT 'exact_match';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS gold_data JSONB;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS primary_metric TEXT NOT NULL DEFAULT 'exact_match';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS secondary_metrics JSONB DEFAULT '[]'::jsonb;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS threshold_pass REAL NOT NULL DEFAULT 0.85;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS threshold_regression_alert REAL NOT NULL DEFAULT 0.90;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS execution_model TEXT NOT NULL DEFAULT 'gpt-4o';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS execution_temperature REAL NOT NULL DEFAULT 0.0;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS execution_max_tokens INTEGER NOT NULL DEFAULT 2048;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS execution_seed BIGINT;

-- Update difficulty column to be nullable (was NOT NULL)
ALTER TABLE tasks ALTER COLUMN difficulty DROP NOT NULL;
