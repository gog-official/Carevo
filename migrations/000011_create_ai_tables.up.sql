CREATE TABLE IF NOT EXISTS survey_questions (
    id          BIGSERIAL PRIMARY KEY,
    category    VARCHAR(50) NOT NULL,
    question_text TEXT      NOT NULL,
    options     JSONB      NOT NULL DEFAULT '[]',
    sort_order  INT        NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS survey_responses (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES survey_questions(id) ON DELETE CASCADE,
    answer      JSONB  NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, question_id)
);

CREATE TABLE IF NOT EXISTS ai_results (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    raw_response       JSONB,
    recommended_careers JSONB,
    status             VARCHAR(20)  NOT NULL DEFAULT 'processing',
    error_message      TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at       TIMESTAMPTZ,
    UNIQUE (user_id)
);

CREATE TABLE IF NOT EXISTS career_scores (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id   BIGINT       NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    score       INT          NOT NULL CHECK (score >= 0 AND score <= 100),
    reasoning   TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, career_id)
);
