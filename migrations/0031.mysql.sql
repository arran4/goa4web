-- Normalize roles and rename permissions table
CREATE TABLE roles (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY roles_name_idx (name)
);

ALTER TABLE permissions ADD COLUMN role_id INT;

INSERT INTO roles (name)
SELECT DISTINCT role FROM permissions;

UPDATE permissions p
JOIN roles r ON p.role = r.name
SET p.role_id = r.id;

ALTER TABLE permissions DROP COLUMN role;

RENAME TABLE permissions TO user_roles;

-- Record upgrade to schema version 31
UPDATE schema_version SET version = 31 WHERE version = 30;
