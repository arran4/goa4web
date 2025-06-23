-- Add expires_at column to userstopiclevel for permission expiration tracking
ALTER TABLE userstopiclevel
    ADD COLUMN IF NOT EXISTS expires_at DATETIME DEFAULT NULL;
