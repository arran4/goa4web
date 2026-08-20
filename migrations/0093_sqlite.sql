-- +goose Up
-- ALTER TABLE external_links MODIFY card_description TEXT;
UPDATE schema_version SET version = 93;