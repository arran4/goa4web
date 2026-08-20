-- +goose Up
UPDATE roles SET public_profile_allowed_at = CURRENT_TIMESTAMP WHERE public_profile_allowed_at IS NULL AND can_login = 1;
UPDATE schema_version SET version = 76;