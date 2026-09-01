-- +goose Up
-- SQLite-only schema repair; MySQL has allowed NULL language IDs since migration 0060.
UPDATE schema_version SET version = 98;

-- +goose Down
UPDATE schema_version SET version = 97;
