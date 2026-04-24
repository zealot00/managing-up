DROP INDEX IF EXISTS idx_mcp_catalog_use_count;
DROP INDEX IF EXISTS idx_mcp_catalog_enabled;
DROP INDEX IF EXISTS idx_mcp_catalog_server_id;
DROP TABLE IF EXISTS mcp_router_catalog;

DROP INDEX IF EXISTS idx_mcp_permissions_skill;
DROP INDEX IF EXISTS idx_mcp_permissions_api_key;
DROP INDEX IF EXISTS idx_mcp_permissions_user;
DROP INDEX IF EXISTS idx_mcp_permissions_server;
DROP TABLE IF EXISTS mcp_server_permissions;