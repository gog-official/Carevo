CREATE INDEX IF NOT EXISTS idx_careers_fts
    ON careers
    USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '') || ' ' || coalesce(summary, '')));
