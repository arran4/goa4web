-- +goose Up
ALTER TABLE users ADD COLUMN public_profile_enabled_at DATETIME DEFAULT NULL AFTER deleted_at;

ALTER TABLE roles ADD COLUMN public_profile_allowed_at DATETIME DEFAULT NULL AFTER is_admin;