-- +goose Up
UPDATE roles SET name = 'anonymous' WHERE name = 'reader';
UPDATE roles SET name = 'user' WHERE name = 'writer';
INSERT INTO roles (name)
SELECT 'content writer' WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'content writer');
UPDATE schema_version SET version = 34;