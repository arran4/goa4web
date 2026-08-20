-- +goose Up
ALTER TABLE user_passkeys ADD COLUMN name TEXT NOT NULL DEFAULT 'Passkey';
UPDATE schema_version SET version = 91;