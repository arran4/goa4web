ALTER TABLE faq ADD COLUMN IF NOT EXISTS updated_at DATETIME DEFAULT NULL;
ALTER TABLE faq_categories ADD COLUMN IF NOT EXISTS updated_at DATETIME DEFAULT NULL;
ALTER TABLE faq_categories ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 0;

-- Record upgrade to schema version 83
UPDATE schema_version SET version = 83 WHERE version = 82;
