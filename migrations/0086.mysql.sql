-- +goose Up
ALTER TABLE preferences ADD COLUMN image_safe_dimension VARCHAR(50);
UPDATE schema_version SET version = 86 WHERE version = 85;

-- +goose Down
ALTER TABLE preferences DROP COLUMN image_safe_dimension;
UPDATE schema_version SET version = 85 WHERE version = 86;
