CREATE TABLE IF NOT EXISTS careers (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL UNIQUE,
    summary         TEXT        NOT NULL DEFAULT '',
    description     TEXT        NOT NULL DEFAULT '',
    category_id     BIGINT      NOT NULL REFERENCES categories(id),
    daily_tasks     JSONB       NOT NULL DEFAULT '[]',
    skills          TEXT[]      NOT NULL DEFAULT '{}',
    salary_min      INT         NOT NULL DEFAULT 0,
    salary_max      INT         NOT NULL DEFAULT 0,
    salary_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    salary_period   VARCHAR(20) NOT NULL DEFAULT 'yearly',
    difficulty      INT         NOT NULL DEFAULT 1 CHECK (difficulty >= 1 AND difficulty <= 5),
    future_proof_score INT     NOT NULL DEFAULT 0 CHECK (future_proof_score >= 0 AND future_proof_score <= 100),
    education_required  TEXT    NOT NULL DEFAULT '',
    outlook         TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_careers_category_id ON careers (category_id);
CREATE INDEX idx_careers_slug ON careers (slug);
