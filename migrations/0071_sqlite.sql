-- +goose Up
ALTER TABLE imageboard ADD COLUMN deleted_at DATETIME DEFAULT NULL;