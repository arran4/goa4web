-- +goose Up
ALTER TABLE preferences ADD COLUMN auto_subscribe_replies TINYINT(1) NOT NULL DEFAULT 1;
UPDATE schema_version SET version = 12;