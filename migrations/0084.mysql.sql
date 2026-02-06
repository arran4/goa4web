ALTER TABLE sessions ADD COLUMN IF NOT EXISTS branch_name varchar(255);

-- Record upgrade to schema version 84
UPDATE schema_version SET version = 84 WHERE version = 83;
