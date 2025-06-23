-- Add expiry and cancellation columns
ALTER TABLE banned_ips ADD COLUMN expires_at datetime DEFAULT NULL;
ALTER TABLE banned_ips ADD COLUMN canceled_at datetime DEFAULT NULL;

-- Record upgrade to schema version 8
UPDATE schema_version SET version = 8 WHERE version = 7;
