-- +goose Up
ALTER TABLE userstopiclevel ADD COLUMN expires_at DATETIME DEFAULT NULL;
UPDATE schema_version SET version = 5;