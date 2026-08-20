-- +goose Up
-- ALTER TABLE user_roles DROP COLUMN section;
UPDATE schema_version SET version = 32;