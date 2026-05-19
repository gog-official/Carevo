CREATE TABLE IF NOT EXISTS bookmarks (
    user_id    BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id  BIGINT    NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, career_id)
);

CREATE TABLE IF NOT EXISTS challenges (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id      BIGINT    NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_streak INT       NOT NULL DEFAULT 0,
    longest_streak INT       NOT NULL DEFAULT 0,
    is_active      BOOLEAN   NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS challenge_checkins (
    id            BIGSERIAL PRIMARY KEY,
    challenge_id  BIGINT       NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    checkin_date  DATE         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (challenge_id, checkin_date)
);

CREATE TABLE IF NOT EXISTS project_ideas (
    id          BIGSERIAL PRIMARY KEY,
    career_id   BIGINT       NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
