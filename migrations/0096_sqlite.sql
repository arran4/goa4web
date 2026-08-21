-- +goose Up
UPDATE schema_version SET version = 96;

-- +goose Down
UPDATE schema_version SET version = 95;
