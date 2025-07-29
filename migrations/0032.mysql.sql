-- Drop obsolete section column from user_roles
ALTER TABLE user_roles DROP COLUMN section;

-- Update schema version
UPDATE schema_version SET version = 32 WHERE version = 31;
