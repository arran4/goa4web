-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_emails MODIFY COLUMN email VARCHAR(255) NOT NULL DEFAULT '';
UPDATE schema_version SET version = 100;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE schema_version SET version = 99;
-- +goose StatementEnd
