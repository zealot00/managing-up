-- MCP Server Permissions: Bind MCP servers to API keys/users
CREATE TABLE mcp_server_permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mcp_server_id   UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    user_id         TEXT,
    api_key_id      TEXT,
    skill_id        TEXT,
    permission_type VARCHAR(50) NOT NULL DEFAULT 'invoke',
    is_granted      BOOLEAN NOT NULL DEFAULT true,
    granted_by      TEXT,
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,
    CONSTRAINT chk_mcp_permission_type CHECK (permission_type IN ('invoke', 'admin', 'read')),
    CONSTRAINT chk_mcp_permission_target CHECK (user_id IS NOT NULL OR api_key_id IS NOT NULL OR skill_id IS NOT NULL)
);

CREATE INDEX idx_mcp_permissions_server ON mcp_server_permissions(mcp_server_id);
CREATE INDEX idx_mcp_permissions_user ON mcp_server_permissions(user_id);
CREATE INDEX idx_mcp_permissions_api_key ON mcp_server_permissions(api_key_id);
CREATE INDEX idx_mcp_permissions_skill ON mcp_server_permissions(skill_id);

-- MCP Router Catalog: Persistent catalog for MCP servers available for routing
CREATE TABLE mcp_router_catalog (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       VARCHAR(255) UNIQUE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    trust_score     FLOAT NOT NULL DEFAULT 0.5,
    transport_type  VARCHAR(50) NOT NULL,
    url             TEXT,
    headers         JSONB DEFAULT '{}',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    approved_by     TEXT,
    metadata_       JSONB DEFAULT '{}',
    use_count      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mcp_catalog_server_id ON mcp_router_catalog(server_id);
CREATE INDEX idx_mcp_catalog_enabled ON mcp_router_catalog(enabled);
CREATE INDEX idx_mcp_catalog_use_count ON mcp_router_catalog(use_count DESC);