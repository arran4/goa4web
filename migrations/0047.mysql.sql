-- Add path and details columns to audit_log
ALTER TABLE audit_log
    ADD COLUMN path text NOT NULL DEFAULT '' AFTER action,
    ADD COLUMN details text AFTER path,
    ADD COLUMN data text AFTER details;

-- Update schema version
UPDATE schema_version SET version = 47 WHERE version = 46;
