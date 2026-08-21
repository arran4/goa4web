-- +goose Up
UPDATE user_emails SET email = '' WHERE email IS NULL;
-- ALTER TABLE user_emails MODIFY COLUMN email tinytext NOT NULL DEFAULT '';
UPDATE schema_version SET version = 70;