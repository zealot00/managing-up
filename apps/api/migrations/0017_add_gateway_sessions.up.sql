-- Gateway Sessions: 会话头，关联所有路由和执行事件
CREATE TABLE mcp_gateway_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_type        VARCHAR(50) NOT NULL DEFAULT 'router',
    agent_id            VARCHAR(255) NOT NULL,
    correlation_id      VARCHAR(255) NOT NULL,
    task_intent         JSONB NOT NULL DEFAULT '{}',
    risk_level          VARCHAR(50) DEFAULT 'low',
    policy_decision     JSONB,
    status              VARCHAR(50) DEFAULT 'active',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at            TIMESTAMPTZ,
    metadata_           JSONB DEFAULT '{}'
);

CREATE INDEX idx_gateway_sessions_correlation ON mcp_gateway_sessions(correlation_id);
CREATE INDEX idx_gateway_sessions_agent ON mcp_gateway_sessions(agent_id);
CREATE INDEX idx_gateway_sessions_status ON mcp_gateway_sessions(status);
CREATE INDEX idx_gateway_sessions_started_at ON mcp_gateway_sessions(started_at DESC);