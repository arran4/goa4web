-- +goose Up
ALTER TABLE writing ADD COLUMN timezone TEXT DEFAULT NULL;
ALTER TABLE deactivated_writings ADD COLUMN timezone TEXT DEFAULT NULL;
UPDATE schema_version SET version = 69;