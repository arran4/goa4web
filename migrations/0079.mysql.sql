-- Record upgrade to schema version 79
UPDATE schema_version SET version = 79 WHERE version = 78;
