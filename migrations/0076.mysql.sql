-- Backfill public_profile_allowed_at for roles that can login
UPDATE roles SET public_profile_allowed_at = NOW() WHERE public_profile_allowed_at IS NULL AND can_login = 1;

-- Update schema version
UPDATE schema_version SET version = 76 WHERE version = 75;
