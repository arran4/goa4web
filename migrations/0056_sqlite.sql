-- +goose Up
ALTER TABLE preferences ADD COLUMN timezone TEXT DEFAULT NULL;
UPDATE schema_version SET version = 56;