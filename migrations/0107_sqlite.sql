-- +goose Up
UPDATE schema_version SET version = 107;

-- +goose Down
UPDATE schema_version SET version = 106;
