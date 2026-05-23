DROP TABLE IF EXISTS career_suggestions;
DROP INDEX IF EXISTS idx_career_suggestions_status;

ALTER TABLE users DROP COLUMN IF EXISTS is_admin;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
