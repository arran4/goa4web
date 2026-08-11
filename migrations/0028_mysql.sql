-- +goose Up
ALTER TABLE user_emails
    ADD COLUMN notification_priority int NOT NULL DEFAULT 0;

-- Set high priority for existing verified emails
UPDATE user_emails SET notification_priority = 100 WHERE verified_at IS NOT NULL;

-- Record upgrade to schema version 28
