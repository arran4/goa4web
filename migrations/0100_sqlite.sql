-- +goose Up
-- +goose StatementBegin
UPDATE schema_version SET version = 100;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE schema_version SET version = 99;
-- +goose StatementEnd
