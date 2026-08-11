-- +goose Up
ALTER TABLE faq ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT '';

-- Record upgrade to schema version 84
