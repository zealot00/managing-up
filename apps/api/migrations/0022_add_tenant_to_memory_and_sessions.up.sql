-- Add tenant isolation to memory_cells
ALTER TABLE memory_cells ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) NOT NULL DEFAULT 'default';
CREATE INDEX IF NOT EXISTS idx_memory_cells_tenant ON memory_cells(tenant_id);

-- Add tenant_id to gateway sessions for multi-tenancy
ALTER TABLE mcp_gateway_sessions ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255) NOT NULL DEFAULT 'default';
CREATE INDEX IF NOT EXISTS idx_gateway_sessions_tenant ON mcp_gateway_sessions(tenant_id);
