ALTER TABLE approvals DROP COLUMN IF EXISTS resolution_note;
ALTER TABLE approvals DROP COLUMN IF EXISTS approved_by;

ALTER TABLE executions DROP COLUMN IF EXISTS updated_at;

ALTER TABLE skill_versions DROP COLUMN IF EXISTS created_by;

ALTER TABLE skills DROP COLUMN IF EXISTS created_by;
