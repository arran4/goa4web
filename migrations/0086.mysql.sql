-- +goose Up
ALTER TABLE preferences ADD COLUMN image_safe_dimension VARCHAR(50);
UPDATE schema_version SET version = 86 WHERE version = 85;
