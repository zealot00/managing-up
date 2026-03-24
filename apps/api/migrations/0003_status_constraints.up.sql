-- Add CHECK constraints for status fields

-- skills.status
ALTER TABLE skills DROP CONSTRAINT IF EXISTS chk_skills_status;
ALTER TABLE skills ADD CONSTRAINT chk_skills_status CHECK (status IN ('draft', 'published', 'deprecated'));

-- skill_versions.status
ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS chk_skill_versions_status;
ALTER TABLE skill_versions ADD CONSTRAINT chk_skill_versions_status CHECK (status IN ('draft', 'published', 'deprecated'));

-- executions.status
ALTER TABLE executions DROP CONSTRAINT IF EXISTS chk_executions_status;
ALTER TABLE executions ADD CONSTRAINT chk_executions_status CHECK (status IN ('pending', 'running', 'waiting_approval', 'succeeded', 'failed', 'stopped'));

-- approvals.status
ALTER TABLE approvals DROP CONSTRAINT IF EXISTS chk_approvals_status;
ALTER TABLE approvals ADD CONSTRAINT chk_approvals_status CHECK (status IN ('waiting', 'approved', 'rejected'));

-- procedure_drafts.validation_status
ALTER TABLE procedure_drafts DROP CONSTRAINT IF EXISTS chk_procedure_drafts_validation_status;
ALTER TABLE procedure_drafts ADD CONSTRAINT chk_procedure_drafts_validation_status CHECK (validation_status IN ('draft', 'validated', 'rejected'));
