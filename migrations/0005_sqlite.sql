-- +goose Up
ALTER TABLE userstopiclevel
    ADD COLUMN IF NOT EXISTS expires_at DATETIME DEFAULT NULL;