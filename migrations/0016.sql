-- Remove obsolete columns from subscriptions table
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS item_type,
    DROP COLUMN IF EXISTS target_id;

-- Record upgrade to schema version 16
UPDATE schema_version SET version = 16 WHERE version = 15;
