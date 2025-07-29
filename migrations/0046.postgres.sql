-- Add login and admin flags to roles
ALTER TABLE roles
    ADD COLUMN can_login TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN is_admin TINYINT(1) NOT NULL DEFAULT 0;

-- Grant login ability to regular roles
UPDATE roles SET can_login = 1 WHERE name IN ('user','content writer','moderator','administrator');

-- Mark administrator role
UPDATE roles SET is_admin = 1 WHERE name = 'administrator';

-- Update schema version
UPDATE schema_version SET version = 46 WHERE version = 45;
