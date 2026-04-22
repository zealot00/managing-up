DROP INDEX IF EXISTS idx_mcp_router_logs_session;
ALTER TABLE mcp_router_logs DROP COLUMN IF EXISTS session_id;
