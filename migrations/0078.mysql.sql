-- Record upgrade to schema version 78
UPDATE schema_version SET version = 78 WHERE version = 77;
