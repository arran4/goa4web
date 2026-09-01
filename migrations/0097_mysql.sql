-- +goose Up
-- Drop obsolete subscription columns if they exist
ALTER TABLE subscriptions DROP COLUMN IF EXISTS item_type;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS target_id;
UPDATE schema_version SET version = 97;

-- +goose Down
UPDATE schema_version SET version = 96;
