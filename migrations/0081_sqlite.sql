-- +goose Up
ALTER TABLE external_links ADD COLUMN card_duration TEXT;
ALTER TABLE external_links ADD COLUMN card_upload_date TEXT;
ALTER TABLE external_links ADD COLUMN card_author TEXT;

UPDATE uploaded_images SET path = ltrim(path, '/');
UPDATE uploaded_images SET path = ltrim(path, 'uploads');
UPDATE uploaded_images SET path = ltrim(path, '/');

UPDATE schema_version SET version = 81;