-- Add approval columns for image BBS uploads
ALTER TABLE imageboard
    ADD COLUMN IF NOT EXISTS approval_required TINYINT(1) NOT NULL DEFAULT 0;
ALTER TABLE imagepost
    ADD COLUMN IF NOT EXISTS approved TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS file_size INT NOT NULL DEFAULT 0;

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
