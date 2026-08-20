-- +goose Up
ALTER TABLE faq ADD COLUMN description TEXT DEFAULT '';
UPDATE schema_version SET version = 84;