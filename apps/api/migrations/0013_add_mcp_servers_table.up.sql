CREATE TABLE IF NOT EXISTS mcp_servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    transport_type TEXT NOT NULL DEFAULT 'stdio',
    command TEXT,
    args JSONB DEFAULT '[]'::jsonb,
    env JSONB DEFAULT '{}'::jsonb,
    url TEXT,
    headers JSONB DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,
    approved_by TEXT,
    approved_at TIMESTAMPTZ,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_mcp_server_status CHECK (status IN ('pending', 'approved', 'rejected', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_mcp_servers_name ON mcp_servers(name);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_status ON mcp_servers(status);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled ON mcp_servers(is_enabled);
