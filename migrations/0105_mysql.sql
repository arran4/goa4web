-- +goose Up
-- MySQL already accepts the current MIME-based queue insert.
UPDATE schema_version SET version = 105;

-- +goose Down
UPDATE schema_version SET version = 104;
