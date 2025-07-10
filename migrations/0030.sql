ALTER TABLE permissions
    CHANGE COLUMN level role tinyblob DEFAULT NULL;

-- Record upgrade to schema version 30
UPDATE schema_version SET version = 30 WHERE version = 29;
