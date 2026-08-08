-- +goose Up
-- Enforce non-null fields on user_roles
ALTER TABLE user_roles
    MODIFY users_idusers INT NOT NULL,
    MODIFY role_id INT NOT NULL;

-- Update schema version
