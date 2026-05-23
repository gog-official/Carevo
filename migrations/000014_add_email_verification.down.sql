DROP TABLE IF EXISTS verification_codes;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
