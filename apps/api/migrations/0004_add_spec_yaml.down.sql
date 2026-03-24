ALTER TABLE skill_versions DROP COLUMN IF EXISTS spec_yaml;
ALTER TABLE executions DROP COLUMN IF EXISTS created_by;
ALTER TABLE executions DROP COLUMN IF EXISTS ended_at;
ALTER TABLE executions DROP COLUMN IF EXISTS duration_ms;
DROP TABLE IF EXISTS execution_steps;
