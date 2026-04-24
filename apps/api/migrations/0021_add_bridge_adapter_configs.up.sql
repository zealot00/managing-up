-- Bridge Adapter: REST API to MCP Adapter Configuration
CREATE TABLE bridge_adapter_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    adapter_type        VARCHAR(50) NOT NULL DEFAULT 'openapi',
    config              JSONB NOT NULL DEFAULT '{}',
    tools               JSONB NOT NULL DEFAULT '[]',
    enabled             BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bridge_adapter_name ON bridge_adapter_configs(name);
CREATE INDEX idx_bridge_adapter_type ON bridge_adapter_configs(adapter_type);
CREATE INDEX idx_bridge_adapter_enabled ON bridge_adapter_configs(enabled);