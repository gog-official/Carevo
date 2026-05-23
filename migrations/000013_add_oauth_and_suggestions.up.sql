ALTER TABLE users ADD COLUMN auth_provider VARCHAR(50) NOT NULL DEFAULT 'email';
ALTER TABLE users ADD COLUMN auth_provider_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_users_auth ON users(auth_provider, auth_provider_id);

ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS career_suggestions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    reason TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_career_suggestions_status ON career_suggestions(status);
