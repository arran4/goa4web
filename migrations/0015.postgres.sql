-- Rename worker_errors table to dead_letters
ALTER TABLE worker_errors RENAME TO dead_letters;

-- Record upgrade to schema version 15
UPDATE schema_version SET version = 15 WHERE version = 14;
