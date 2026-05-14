CREATE TABLE llm_fallback_chains (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model       TEXT NOT NULL,
    is_enabled  BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(model)
);

CREATE TABLE llm_fallback_targets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id    UUID NOT NULL REFERENCES llm_fallback_chains(id) ON DELETE CASCADE,
    provider    TEXT NOT NULL,
    model       TEXT NOT NULL,
    weight      INT NOT NULL DEFAULT 100,
    priority    INT NOT NULL DEFAULT 0,
    is_enabled  BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_fallback_targets_chain_id ON llm_fallback_targets(chain_id);
