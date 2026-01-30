ALTER TABLE faq_categories ADD COLUMN deleted_at DATETIME DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 81 WHERE version = 80;
