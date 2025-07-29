-- Add auto_subscribe_replies column to preferences
ALTER TABLE preferences
    ADD COLUMN IF NOT EXISTS auto_subscribe_replies TINYINT(1) NOT NULL DEFAULT 1;

-- Record upgrade to schema version 12
UPDATE schema_version SET version = 12 WHERE version = 11;
