-- +goose Up
-- Drop obsolete section column from user_roles
ALTER TABLE user_roles DROP COLUMN section;

-- Update schema version
