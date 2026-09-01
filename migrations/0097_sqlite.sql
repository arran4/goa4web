-- +goose Up
-- Drop obsolete subscription columns from SQLite
ALTER TABLE subscriptions DROP COLUMN item_type;
ALTER TABLE subscriptions DROP COLUMN target_id;
UPDATE schema_version SET version = 97;

-- +goose Down
UPDATE schema_version SET version = 96;
