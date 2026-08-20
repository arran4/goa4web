-- +goose Up
ALTER TABLE user_emails ADD COLUMN verification_expires_at datetime DEFAULT NULL;
UPDATE schema_version SET version = 29;