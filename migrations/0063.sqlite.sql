ALTER TABLE comments ADD COLUMN timezone TEXT;
ALTER TABLE deactivated_comments ADD COLUMN timezone TEXT;

UPDATE schema_version SET version = 63 WHERE version = 62;

