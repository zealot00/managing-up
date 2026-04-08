-- MCP Router Catalog
CREATE TABLE mcp_router_catalog (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       TEXT NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    transport_type  VARCHAR(50) NOT NULL,
    command         VARCHAR(500),
    args            TEXT[],
    url             VARCHAR(500),
    task_types      TEXT[] NOT NULL,
    tags            TEXT[],
    capabilities    JSONB NOT NULL DEFAULT '{}',
    routing_config  JSONB DEFAULT '{}',
    status          VARCHAR(50) DEFAULT 'active',
    trust_score     DECIMAL(3,2) DEFAULT 0.5,
    use_count       BIGINT DEFAULT 0,
    error_count     BIGINT DEFAULT 0,
    last_used_at    TIMESTAMP,
    synced_at       TIMESTAMP DEFAULT NOW(),
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_catalog_transport CHECK (transport_type IN ('stdio', 'http')),
    CONSTRAINT chk_mcp_catalog_status CHECK (status IN ('active', 'disabled', 'maintenance')),
    CONSTRAINT chk_mcp_catalog_trust CHECK (trust_score >= 0 AND trust_score <= 1)
);

CREATE INDEX idx_mcp_catalog_task_types ON mcp_router_catalog USING GIN(task_types);
CREATE INDEX idx_mcp_catalog_tags ON mcp_router_catalog USING GIN(tags);
CREATE INDEX idx_mcp_catalog_status ON mcp_router_catalog(status);
CREATE INDEX idx_mcp_catalog_trust ON mcp_router_catalog(trust_score DESC);

-- MCP Router Logs
CREATE TABLE mcp_router_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id      VARCHAR(255) NOT NULL,
    agent_id            VARCHAR(255),
    task_type           VARCHAR(100),
    task_tags           TEXT[],
    task_complexity     VARCHAR(50),
    raw_description     TEXT,
    matched             BOOLEAN NOT NULL,
    matched_server_id   TEXT,
    match_score         DECIMAL(3,2),
    match_latency_ms    INTEGER,
    status              VARCHAR(50),
    error_code          VARCHAR(50),
    error_message       TEXT,
    duration_ms         INTEGER,
    created_at          TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_log_status CHECK (status IN ('success', 'failure', 'timeout'))
);

CREATE INDEX idx_mcp_router_logs_created ON mcp_router_logs(created_at);
CREATE INDEX idx_mcp_router_logs_agent ON mcp_router_logs(agent_id);
CREATE INDEX idx_mcp_router_logs_correlation ON mcp_router_logs(correlation_id);
CREATE INDEX idx_mcp_router_logs_server ON mcp_router_logs(matched_server_id);

-- MCP Router Sync Log
CREATE TABLE mcp_router_sync_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       TEXT NOT NULL,
    sync_type       VARCHAR(50) NOT NULL,
    old_status      VARCHAR(50),
    new_status      VARCHAR(50),
    approved_by     TEXT,
    approved_at     TIMESTAMP,
    note            TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_mcp_sync_type CHECK (sync_type IN ('approved_sync', 'status_change', 'removal', 'catalog_update'))
);

CREATE INDEX idx_mcp_sync_log_server ON mcp_router_sync_log(server_id);
