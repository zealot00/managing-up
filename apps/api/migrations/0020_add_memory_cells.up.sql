-- Memory Cells: 存储 Agent 的记忆单元
CREATE TABLE memory_cells (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope               VARCHAR(50) NOT NULL DEFAULT 'session',
    agent_id            VARCHAR(255) NOT NULL,
    session_id          UUID REFERENCES mcp_gateway_sessions(id) ON DELETE SET NULL,
    execution_id        TEXT,
    key                 VARCHAR(255) NOT NULL,
    value               JSONB NOT NULL,
    value_type          VARCHAR(50) NOT NULL DEFAULT 'text',
    metadata_           JSONB DEFAULT '{}',
    tags                TEXT[],
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    CONSTRAINT chk_memory_scope CHECK (scope IN ('session', 'execution', 'agent', 'tenant')),
    CONSTRAINT chk_memory_value_type CHECK (value_type IN ('text', 'json', 'binary'))
);

CREATE INDEX idx_memory_cells_scope ON memory_cells(scope);
CREATE INDEX idx_memory_cells_agent ON memory_cells(agent_id);
CREATE INDEX idx_memory_cells_session ON memory_cells(session_id);
CREATE INDEX idx_memory_cells_execution ON memory_cells(execution_id);
CREATE INDEX idx_memory_cells_key ON memory_cells(key);
CREATE INDEX idx_memory_cells_created ON memory_cells(created_at DESC);