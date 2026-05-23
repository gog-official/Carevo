-- Add expanded career fields for Nepal-specific salary tiers, demand data, metadata
ALTER TABLE careers ADD COLUMN IF NOT EXISTS salary_tiers JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS city_salaries JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS demand_data JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS source_labels JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS ne_content JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS career_metadata JSONB DEFAULT '{}'::jsonb;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS work_life_balance INT NOT NULL DEFAULT 3 CHECK (work_life_balance >= 1 AND work_life_balance <= 5);
ALTER TABLE careers ADD COLUMN IF NOT EXISTS study_duration VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE careers ADD COLUMN IF NOT EXISTS degree_required VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE careers ADD COLUMN IF NOT EXISTS creative_score INT NOT NULL DEFAULT 3 CHECK (creative_score >= 1 AND creative_score <= 5);
ALTER TABLE careers ADD COLUMN IF NOT EXISTS technical_score INT NOT NULL DEFAULT 3 CHECK (technical_score >= 1 AND technical_score <= 5);
ALTER TABLE careers ADD COLUMN IF NOT EXISTS freelance_potential INT NOT NULL DEFAULT 1 CHECK (freelance_potential >= 1 AND freelance_potential <= 5);
ALTER TABLE careers ADD COLUMN IF NOT EXISTS is_government BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE careers ADD COLUMN IF NOT EXISTS is_remote_ok BOOLEAN NOT NULL DEFAULT FALSE;

DO $$ BEGIN
    ALTER TABLE careers ADD COLUMN IF NOT EXISTS exam_required VARCHAR(255) NOT NULL DEFAULT '';
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

-- Progress tracker tables
CREATE TABLE IF NOT EXISTS user_skills (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    skill_name VARCHAR(200) NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, career_id, skill_name)
);

CREATE TABLE IF NOT EXISTS user_certifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    cert_name VARCHAR(200) NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, career_id, cert_name)
);

CREATE TABLE IF NOT EXISTS user_goals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    goal_text TEXT NOT NULL,
    goal_type VARCHAR(20) NOT NULL DEFAULT 'daily' CHECK (goal_type IN ('daily', 'weekly', 'milestone')),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    due_date DATE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_milestones (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    milestone_name VARCHAR(255) NOT NULL,
    is_reached BOOLEAN NOT NULL DEFAULT FALSE,
    reached_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_study_streaks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    study_date DATE NOT NULL,
    minutes_studied INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, study_date)
);

CREATE TABLE IF NOT EXISTS user_job_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL DEFAULT '',
    position_name VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'saved' CHECK (status IN ('saved', 'applied', 'interview', 'offer', 'rejected', 'accepted')),
    notes TEXT NOT NULL DEFAULT '',
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_mode_preferences (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (mode IN ('student', 'parent')),
    language VARCHAR(10) NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'ne')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Career comparison table
CREATE TABLE IF NOT EXISTS career_comparisons (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    career_ids BIGINT[] NOT NULL DEFAULT '{}',
    name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Verified source data table
CREATE TABLE IF NOT EXISTS career_sources (
    id BIGSERIAL PRIMARY KEY,
    career_id BIGINT NOT NULL REFERENCES careers(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    source_url TEXT NOT NULL DEFAULT '',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    confidence_score INT NOT NULL DEFAULT 0 CHECK (confidence_score >= 0 AND confidence_score <= 100),
    last_checked TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_career_sources_career_id ON career_sources(career_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_user_career ON user_skills(user_id, career_id);
CREATE INDEX IF NOT EXISTS idx_user_goals_user_career ON user_goals(user_id, career_id);
CREATE INDEX IF NOT EXISTS idx_user_study_streaks_user ON user_study_streaks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_job_apps_user ON user_job_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_career_comparisons_user ON career_comparisons(user_id);
CREATE INDEX IF NOT EXISTS idx_careers_government ON careers(is_government);
CREATE INDEX IF NOT EXISTS idx_careers_remote ON careers(is_remote_ok);
