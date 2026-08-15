-- +goose Up
-- ALTER TABLE user_roles
    MODIFY users_idusers INT NOT NULL,
    MODIFY role_id INT NOT NULL;