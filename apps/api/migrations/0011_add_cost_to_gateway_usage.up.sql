DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'gateway_usage_events' AND column_name = 'cost') THEN
        ALTER TABLE gateway_usage_events ADD COLUMN cost DECIMAL(10, 6) NOT NULL DEFAULT 0;
    END IF;
END $$;
