CREATE TABLE IF NOT EXISTS tips (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    author TEXT,
    category TEXT NOT NULL DEFAULT 'quote',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tips_active ON tips(is_active);
CREATE INDEX IF NOT EXISTS idx_tips_category ON tips(category);
