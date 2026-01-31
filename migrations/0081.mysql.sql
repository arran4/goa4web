-- Record upgrade to schema version 81
UPDATE schema_version SET version = 81 WHERE version = 80;
