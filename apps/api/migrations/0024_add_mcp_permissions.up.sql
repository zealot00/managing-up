-- MCP Server Permissions: Bind MCP servers to API keys/users
-- mcp_servers.id is TEXT, so mcp_server_id must be TEXT not UUID
CREATE TABLE IF NOT EXISTS mcp_server_permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mcp_server_id   TEXT NOT NULL,
    user_id         TEXT,
    api_key_id      TEXT,
    skill_id        TEXT,
    permission_type VARCHAR(50) NOT NULL DEFAULT 'invoke',
    is_granted      BOOLEAN NOT NULL DEFAULT true,
    granted_by      TEXT,
    granted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

-- Fix column type if table was created with UUID mcp_server_id in a prior partial run
DO $$ BEGIN
    ALTER TABLE mcp_server_permissions ALTER COLUMN mcp_server_id TYPE TEXT;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Add constraints separately for idempotency
DO $$ BEGIN
    ALTER TABLE mcp_server_permissions ADD CONSTRAINT fk_mcp_perms_server FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;
DO $$ BEGIN
    ALTER TABLE mcp_server_permissions ADD CONSTRAINT chk_mcp_permission_type CHECK (permission_type IN ('invoke', 'admin', 'read'));
EXCEPTION WHEN OTHERS THEN NULL;
END $$;
DO $$ BEGIN
    ALTER TABLE mcp_server_permissions ADD CONSTRAINT chk_mcp_permission_target CHECK (user_id IS NOT NULL OR api_key_id IS NOT NULL OR skill_id IS NOT NULL);
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_mcp_permissions_server ON mcp_server_permissions(mcp_server_id);
CREATE INDEX IF NOT EXISTS idx_mcp_permissions_user ON mcp_server_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_mcp_permissions_api_key ON mcp_server_permissions(api_key_id);
CREATE INDEX IF NOT EXISTS idx_mcp_permissions_skill ON mcp_server_permissions(skill_id);

-- MCP Router Catalog: Add columns for persistent catalog (table created in 0015)
-- Only add columns that don't exist in the 0015 schema
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'mcp_router_catalog' AND column_name = 'headers') THEN
        ALTER TABLE mcp_router_catalog ADD COLUMN headers JSONB DEFAULT '{}';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'mcp_router_catalog' AND column_name = 'approved_by') THEN
        ALTER TABLE mcp_router_catalog ADD COLUMN approved_by TEXT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_mcp_catalog_server_id ON mcp_router_catalog(server_id);
-- 0015 schema has "status" not "enabled"; skip enabled index if column absent
DO $$ BEGIN CREATE INDEX IF NOT EXISTS idx_mcp_catalog_enabled ON mcp_router_catalog(enabled); EXCEPTION WHEN undefined_column THEN NULL; END $$;
CREATE INDEX IF NOT EXISTS idx_mcp_catalog_use_count ON mcp_router_catalog(use_count DESC);
