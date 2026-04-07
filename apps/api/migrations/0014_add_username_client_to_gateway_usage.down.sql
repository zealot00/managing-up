ALTER TABLE gateway_usage_events DROP COLUMN username;
ALTER TABLE gateway_usage_events DROP COLUMN client_name;
DROP INDEX IF EXISTS idx_gateway_usage_events_client_name;
