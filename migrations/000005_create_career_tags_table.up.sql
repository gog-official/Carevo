CREATE TABLE IF NOT EXISTS career_tags (
    id          BIGSERIAL PRIMARY KEY,
    career_id   BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    tag         VARCHAR(100) NOT NULL,
    UNIQUE(career_id, tag)
);

CREATE INDEX idx_career_tags_career_id ON career_tags (career_id);
CREATE INDEX idx_career_tags_tag ON career_tags (tag);
