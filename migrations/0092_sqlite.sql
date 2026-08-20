-- +goose Up
ALTER TABLE user_passkeys ADD COLUMN backup_eligible BOOLEAN DEFAULT NULL;
ALTER TABLE user_passkeys ADD COLUMN backup_state BOOLEAN DEFAULT NULL;
UPDATE schema_version SET version = 92;