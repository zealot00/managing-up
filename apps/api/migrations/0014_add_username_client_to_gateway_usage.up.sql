DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'gateway_usage_events' AND column_name = 'username') THEN
        ALTER TABLE gateway_usage_events ADD COLUMN username TEXT NOT NULL DEFAULT '';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'gateway_usage_events' AND column_name = 'client_name') THEN
        ALTER TABLE gateway_usage_events ADD COLUMN client_name TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_gateway_usage_events_client_name ON gateway_usage_events(client_name, created_at DESC);
