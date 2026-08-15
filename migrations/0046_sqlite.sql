-- +goose Up
ALTER TABLE roles
    ADD COLUMN can_login TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN is_admin TINYINT(1) NOT NULL DEFAULT 0;

UPDATE roles SET can_login = 1 WHERE name IN ('user','content writer','moderator','administrator');

UPDATE roles SET is_admin = 1 WHERE name = 'administrator';