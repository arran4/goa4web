ALTER TABLE comments ADD COLUMN timezone TINYTEXT DEFAULT NULL;
ALTER TABLE deactivated_comments ADD COLUMN timezone TINYTEXT DEFAULT NULL;

UPDATE schema_version SET version = 63 WHERE version = 62;

