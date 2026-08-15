-- +goose Up
CREATE TABLE roles (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
UNIQUE (name)
);

ALTER TABLE permissions ADD COLUMN role_id INT;

INSERT INTO roles (name)
SELECT DISTINCT role FROM permissions;

UPDATE permissions p
JOIN roles r ON p.role = r.name
SET p.role_id = r.id;

-- ALTER TABLE permissions DROP COLUMN role;

RENAME TABLE permissions TO user_roles;