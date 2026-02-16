ALTER TABLE faq ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT '';

-- Record upgrade to schema version 84
UPDATE schema_version SET version = 84 WHERE version = 83;
