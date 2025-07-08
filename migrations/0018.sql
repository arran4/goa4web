-- Rename userlang table to user_language
ALTER TABLE userlang RENAME TO user_language;

-- Record upgrade to schema version 18
UPDATE schema_version SET version = 18 WHERE version = 17;
