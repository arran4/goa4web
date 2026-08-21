-- +goose Up
ALTER TABLE users ADD COLUMN public_profile_enabled_at DATETIME DEFAULT NULL;
ALTER TABLE roles ADD COLUMN public_profile_allowed_at DATETIME DEFAULT NULL;
UPDATE schema_version SET version = 49;