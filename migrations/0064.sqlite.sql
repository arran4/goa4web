ALTER TABLE blogs ADD COLUMN timezone TEXT;
ALTER TABLE deactivated_blogs ADD COLUMN timezone TEXT;
ALTER TABLE site_news ADD COLUMN timezone TEXT;
ALTER TABLE imagepost ADD COLUMN timezone TEXT;
ALTER TABLE deactivated_imageposts ADD COLUMN timezone TEXT;
ALTER TABLE linker ADD COLUMN timezone TEXT;
ALTER TABLE linker_queue ADD COLUMN timezone TEXT;
ALTER TABLE deactivated_linker ADD COLUMN timezone TEXT;
ALTER TABLE faq_revisions ADD COLUMN timezone TEXT;

UPDATE schema_version SET version = 64 WHERE version = 63;

