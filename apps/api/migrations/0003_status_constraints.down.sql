-- Remove CHECK constraints for status fields

ALTER TABLE skills DROP CONSTRAINT IF EXISTS chk_skills_status;
ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS chk_skill_versions_status;
ALTER TABLE executions DROP CONSTRAINT IF EXISTS chk_executions_status;
ALTER TABLE approvals DROP CONSTRAINT IF EXISTS chk_approvals_status;
ALTER TABLE procedure_drafts DROP CONSTRAINT IF EXISTS chk_procedure_drafts_validation_status;
