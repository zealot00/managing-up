-- Policy Versioning: Support for policy history and versioning
CREATE TABLE policy_versions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    version     VARCHAR(50) NOT NULL,
    description TEXT,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_policy_version UNIQUE (name, version)
);

CREATE TABLE policy_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id   UUID NOT NULL REFERENCES policy_versions(id) ON DELETE CASCADE,
    rule_id     VARCHAR(100) NOT NULL,
    version     VARCHAR(50) NOT NULL DEFAULT 'v1',
    condition   TEXT NOT NULL,
    action      VARCHAR(50) NOT NULL,
    reason      TEXT,
    priority    INTEGER NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_policy_rule UNIQUE (policy_id, rule_id)
);

CREATE INDEX idx_policy_versions_name ON policy_versions(name);
CREATE INDEX idx_policy_versions_default ON policy_versions(is_default);
CREATE INDEX idx_policy_rules_policy ON policy_rules(policy_id);
CREATE INDEX idx_policy_rules_active ON policy_rules(is_active);