ALTER TABLE comments ADD COLUMN timezone TEXT DEFAULT NULL;
ALTER TABLE deactivated_comments ADD COLUMN timezone TEXT DEFAULT NULL;

UPDATE schema_version SET version = 63 WHERE version = 62;

