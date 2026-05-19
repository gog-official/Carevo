CREATE TABLE IF NOT EXISTS roadmap_steps (
    id          BIGSERIAL PRIMARY KEY,
    career_id   BIGINT       NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    step_number INT          NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    duration    VARCHAR(100) NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(career_id, step_number)
);

CREATE INDEX idx_roadmap_steps_career_id ON roadmap_steps (career_id);
