-- +goose Up
ALTER TABLE preferences ADD COLUMN custom_css TEXT;
UPDATE schema_version SET version = 75;