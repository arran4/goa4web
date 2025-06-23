-- Add passwd_algorithm column to track the password hashing scheme
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS passwd_algorithm TINYTEXT DEFAULT NULL;

-- Migrate existing users to the legacy md5 algorithm
UPDATE users SET passwd_algorithm = 'md5' WHERE passwd_algorithm IS NULL;

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
