-- +goose Up
UPDATE roles SET public_profile_allowed_at = NOW() WHERE public_profile_allowed_at IS NULL AND can_login = 1;