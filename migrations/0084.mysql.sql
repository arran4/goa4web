ALTER TABLE faq ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT '';
ALTER TABLE faq ADD COLUMN IF NOT EXISTS version VARCHAR(50) DEFAULT '';

-- Record upgrade to schema version 84
UPDATE schema_version SET version = 84 WHERE version = 83;
