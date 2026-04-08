-- Extend skills table
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_id VARCHAR(100);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_name VARCHAR(255);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_version VARCHAR(50);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS sop_section VARCHAR(255);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS compliance_required BOOLEAN DEFAULT false;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE skills ADD COLUMN IF NOT EXISTS trust_score DECIMAL(3,2) DEFAULT 0.5;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT false;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS published_by TEXT;
ALTER TABLE skills ADD COLUMN IF NOT EXISTS draft_source VARCHAR(50) DEFAULT 'manual';
ALTER TABLE skills ADD COLUMN IF NOT EXISTS draft_source_meta JSONB DEFAULT '{}';
ALTER TABLE skills ADD COLUMN IF NOT EXISTS created_by TEXT;

ALTER TABLE skills ADD CONSTRAINT fk_skills_published_by FOREIGN KEY (published_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD CONSTRAINT fk_skills_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE skills ADD CONSTRAINT chk_skills_trust CHECK (trust_score >= 0 AND trust_score <= 1);

CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_tags ON skills USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_skills_trust ON skills(trust_score DESC);
CREATE INDEX IF NOT EXISTS idx_skills_sop ON skills(sop_id);

-- Extend skill_versions table
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS changelog TEXT;
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS sop_version VARCHAR(50);
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS approved_by TEXT;
ALTER TABLE skill_versions ADD CONSTRAINT fk_skill_versions_approved_by FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_skill_versions_skill ON skill_versions(skill_id);

-- Skill Dependencies
CREATE TABLE skill_dependencies (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id            UUID NOT NULL,
    dependency_skill_id UUID NOT NULL,
    version_constraint   VARCHAR(100) NOT NULL,
    created_at          TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_deps_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_deps_dep FOREIGN KEY (dependency_skill_id) REFERENCES skills(id) ON DELETE RESTRICT,
    UNIQUE(skill_id, dependency_skill_id),
    CONSTRAINT chk_skill_deps_no_self CHECK (skill_id != dependency_skill_id)
);

CREATE INDEX idx_skill_deps_skill ON skill_dependencies(skill_id);
CREATE INDEX idx_skill_deps_dep ON skill_dependencies(dependency_skill_id);

-- Skill Ratings
CREATE TABLE skill_ratings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    UUID NOT NULL,
    user_id     TEXT NOT NULL,
    rating      INTEGER NOT NULL,
    comment     TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_ratings_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_ratings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(skill_id, user_id),
    CONSTRAINT chk_skill_ratings CHECK (rating >= 1 AND rating <= 5)
);

CREATE INDEX idx_skill_ratings_skill ON skill_ratings(skill_id);
CREATE INDEX idx_skill_ratings_user ON skill_ratings(user_id);

-- Skill Installs
CREATE TABLE skill_installs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id        UUID NOT NULL,
    user_id         TEXT,
    version         VARCHAR(50) NOT NULL,
    environment     VARCHAR(50) DEFAULT 'production',
    installed_at    TIMESTAMP DEFAULT NOW(),
    skill_snapshot  JSONB,
    CONSTRAINT fk_skill_installs_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_installs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_skill_installs_env CHECK (environment IN ('production', 'staging', 'development'))
);

CREATE INDEX idx_skill_installs_skill ON skill_installs(skill_id);
CREATE INDEX idx_skill_installs_user ON skill_installs(user_id);
CREATE INDEX idx_skill_installs_env ON skill_installs(environment);

-- Skill Publish Approvals
CREATE TABLE skill_publish_approvals (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id                UUID NOT NULL,
    version                 VARCHAR(50) NOT NULL,
    status                  VARCHAR(50) DEFAULT 'pending',
    submitted_by            TEXT NOT NULL,
    submitted_at            TIMESTAMP DEFAULT NOW(),
    reviewed_by             TEXT,
    reviewed_at             TIMESTAMP,
    review_note             TEXT,
    compliance_check_passed  BOOLEAN DEFAULT false,
    compliance_check_note   TEXT,
    created_at              TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_skill_pub_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_pub_submitted FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_pub_reviewed FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_skill_pub_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_skill_pub_skill ON skill_publish_approvals(skill_id);
CREATE INDEX idx_skill_pub_status ON skill_publish_approvals(status);
CREATE INDEX idx_skill_pub_submitted ON skill_publish_approvals(submitted_by);
