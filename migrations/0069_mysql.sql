-- +goose Up
ALTER TABLE writing ADD COLUMN timezone text DEFAULT NULL;
ALTER TABLE deactivated_writings ADD COLUMN timezone text DEFAULT NULL;

-- Update schema version
