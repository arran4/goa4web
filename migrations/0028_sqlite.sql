-- +goose Up
ALTER TABLE user_emails
    ADD COLUMN notification_priority int NOT NULL DEFAULT 0;

UPDATE user_emails SET notification_priority = 100 WHERE verified_at IS NOT NULL;