ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_reviewed;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_submitted;
ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_skill;
ALTER TABLE skill_publish_approvals DROP INDEX IF EXISTS idx_skill_pub_submitted;
ALTER TABLE skill_publish_approvals DROP INDEX IF EXISTS idx_skill_pub_status;
ALTER TABLE skill_publish_approvals DROP INDEX IF EXISTS idx_skill_pub_skill;
DROP TABLE IF EXISTS skill_publish_approvals;

ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_user;
ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_skill;
DROP INDEX IF EXISTS idx_skill_installs_env;
DROP INDEX IF EXISTS idx_skill_installs_user;
DROP INDEX IF EXISTS idx_skill_installs_skill;
DROP TABLE IF EXISTS skill_installs;

ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_user;
ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_skill;
DROP INDEX IF EXISTS idx_skill_ratings_user;
DROP INDEX IF EXISTS idx_skill_ratings_skill;
DROP TABLE IF EXISTS skill_ratings;

ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_dep;
ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_skill;
DROP INDEX IF EXISTS idx_skill_deps_dep;
DROP INDEX IF EXISTS idx_skill_deps_skill;
DROP TABLE IF EXISTS skill_dependencies;

ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS fk_skill_versions_approved_by;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS approved_by;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS sop_version;
ALTER TABLE skill_versions DROP COLUMN IF EXISTS changelog;

ALTER TABLE skills DROP CONSTRAINT IF EXISTS fk_skills_created_by;
ALTER TABLE skills DROP CONSTRAINT IF EXISTS fk_skills_published_by;
ALTER TABLE skills DROP COLUMN IF EXISTS created_by;
ALTER TABLE skills DROP COLUMN IF EXISTS draft_source_meta;
ALTER TABLE skills DROP COLUMN IF EXISTS draft_source;
ALTER TABLE skills DROP COLUMN IF EXISTS published_by;
ALTER TABLE skills DROP COLUMN IF EXISTS published_at;
ALTER TABLE skills DROP COLUMN IF EXISTS verified;
ALTER TABLE skills DROP COLUMN IF EXISTS trust_score;
ALTER TABLE skills DROP COLUMN IF EXISTS tags;
ALTER TABLE skills DROP COLUMN IF EXISTS category;
ALTER TABLE skills DROP COLUMN IF EXISTS compliance_required;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_section;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_version;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_name;
ALTER TABLE skills DROP COLUMN IF EXISTS sop_id;
