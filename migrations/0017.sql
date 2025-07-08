-- Rename occured column to occurred on siteNews table
ALTER TABLE siteNews CHANGE COLUMN occured occurred datetime DEFAULT NULL;

-- Record upgrade to schema version 17
UPDATE schema_version SET version = 17 WHERE version = 16;
