-- Add last_index columns to searchable tables
ALTER TABLE comments ADD COLUMN last_index datetime DEFAULT NULL;
ALTER TABLE site_news ADD COLUMN last_index datetime DEFAULT NULL;
ALTER TABLE blogs ADD COLUMN last_index datetime DEFAULT NULL;
ALTER TABLE writing ADD COLUMN last_index datetime DEFAULT NULL;
ALTER TABLE linker ADD COLUMN last_index datetime DEFAULT NULL;
ALTER TABLE imagepost ADD COLUMN last_index datetime DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 41 WHERE version = 40;
