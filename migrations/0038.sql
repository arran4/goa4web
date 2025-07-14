-- Drop obsolete writing_user_permissions table
DROP TABLE IF EXISTS writing_user_permissions;

-- Update schema version
UPDATE schema_version SET version = 38 WHERE version = 37;
