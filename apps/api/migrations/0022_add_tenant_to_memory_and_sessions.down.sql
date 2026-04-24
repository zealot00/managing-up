DROP INDEX IF EXISTS idx_memory_cells_tenant;
ALTER TABLE memory_cells DROP COLUMN tenant_id;
DROP INDEX IF EXISTS idx_gateway_sessions_tenant;
ALTER TABLE mcp_gateway_sessions DROP COLUMN tenant_id;