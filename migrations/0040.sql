-- Drop legacy writing_user_permissions table
DROP TABLE IF EXISTS writing_user_permissions;

-- Update schema version
UPDATE schema_version SET version = 40 WHERE version = 39;
