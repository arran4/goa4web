-- +goose Up
ALTER TABLE preferences
    ADD COLUMN IF NOT EXISTS auto_subscribe_replies TINYINT(1) NOT NULL DEFAULT 1;