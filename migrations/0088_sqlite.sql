-- +goose Up
ALTER TABLE image_cache_entries ADD COLUMN uploaded_image_id INT DEFAULT NULL;
CREATE INDEX IF NOT EXISTS image_cache_entries_uploaded_image_idx ON image_cache_entries (uploaded_image_id);
UPDATE schema_version SET version = 88;