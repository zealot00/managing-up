CREATE TABLE IF NOT EXISTS gateway_api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gateway_api_keys_user_id ON gateway_api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_gateway_api_keys_active_hash ON gateway_api_keys(key_hash) WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS gateway_usage_events (
    id TEXT PRIMARY KEY,
    api_key_id TEXT NOT NULL REFERENCES gateway_api_keys(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gateway_usage_events_user_id_created_at ON gateway_usage_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gateway_usage_events_created_at ON gateway_usage_events(created_at DESC);
