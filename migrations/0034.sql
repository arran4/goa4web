-- Rename default roles and add new role
UPDATE roles SET name = 'anonymous' WHERE name = 'reader';
UPDATE roles SET name = 'normal user' WHERE name = 'writer';
INSERT INTO roles (name)
SELECT 'content writer' WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'content writer');

-- Update schema version
UPDATE schema_version SET version = 34 WHERE version = 33;
