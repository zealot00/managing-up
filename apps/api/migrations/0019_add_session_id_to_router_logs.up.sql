ALTER TABLE mcp_router_logs ADD COLUMN session_id UUID REFERENCES mcp_gateway_sessions(id) ON DELETE SET NULL;
CREATE INDEX idx_mcp_router_logs_session ON mcp_router_logs(session_id);
