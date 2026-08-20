-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
UNIQUE (name)
);

ALTER TABLE permissions ADD COLUMN role_id INT;

INSERT INTO roles (name)
SELECT DISTINCT role FROM permissions WHERE role IS NOT NULL;

UPDATE permissions
SET role_id = (SELECT id FROM roles WHERE roles.name = permissions.role);

ALTER TABLE permissions RENAME TO user_roles;
UPDATE schema_version SET version = 31;