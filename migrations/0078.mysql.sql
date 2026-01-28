-- Null migration to replace reverted changes
-- This migration only updates the schema version, effectively skipping the grants.
UPDATE schema_version SET version = 78 WHERE version = 77;
