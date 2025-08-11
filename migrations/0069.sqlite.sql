ALTER TABLE writing ADD COLUMN timezone text;
ALTER TABLE deactivated_writings ADD COLUMN timezone text;

-- Update schema version
UPDATE schema_version SET version = 69 WHERE version = 68;
