-- +goose Up
ALTER TABLE faq ADD COLUMN updated_at DATETIME DEFAULT NULL;
ALTER TABLE faq_categories ADD COLUMN updated_at DATETIME DEFAULT NULL;
ALTER TABLE faq_categories ADD COLUMN priority INT NOT NULL DEFAULT 0;
UPDATE schema_version SET version = 83;