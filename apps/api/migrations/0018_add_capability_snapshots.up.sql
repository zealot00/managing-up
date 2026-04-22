-- Skill Capability Snapshots: 记录技能版本的能力评测快照
CREATE TABLE skill_capability_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id            TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version             VARCHAR(50) NOT NULL,
    snapshot_type       VARCHAR(50) NOT NULL DEFAULT 'regression_gate',
    dataset_id          TEXT,
    run_id              TEXT,
    metrics             JSONB NOT NULL DEFAULT '{}',
    overall_score       DECIMAL(5,2),
    passed              BOOLEAN NOT NULL DEFAULT false,
    gate_policy_id      TEXT,
    evaluated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_snapshots_skill ON skill_capability_snapshots(skill_id, version DESC);
CREATE INDEX idx_snapshots_passed ON skill_capability_snapshots(passed);
CREATE INDEX idx_snapshots_evaluated ON skill_capability_snapshots(evaluated_at DESC);
