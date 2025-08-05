-- Normalize writing section name
UPDATE grants SET section = 'writing' WHERE section = 'writings';

-- Update schema version
UPDATE schema_version SET version = 53 WHERE version = 52;
