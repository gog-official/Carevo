CREATE TABLE IF NOT EXISTS resources (
    id          BIGSERIAL PRIMARY KEY,
    career_id   BIGINT       NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    url         TEXT         NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    is_free     BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resources_career_id ON resources (career_id);
