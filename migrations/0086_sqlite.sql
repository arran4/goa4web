-- +goose Up
ALTER TABLE preferences ADD COLUMN image_safe_dimension TEXT;
UPDATE schema_version SET version = 86;