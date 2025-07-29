-- Add expires_at column to userstopiclevel for permission expiration tracking
ALTER TABLE userstopiclevel
    ADD COLUMN IF NOT EXISTS expires_at DATETIME DEFAULT NULL;

-- Record upgrade to schema version 5
UPDATE schema_version SET version = 5 WHERE version = 4;
