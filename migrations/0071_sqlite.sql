-- +goose Up
-- imageboard.deleted_at already added in 0010
UPDATE schema_version SET version = 71;