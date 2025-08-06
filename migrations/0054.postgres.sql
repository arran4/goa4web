-- Add language columns to forum categories and topics
ALTER TABLE forumcategory
    ADD COLUMN IF NOT EXISTS language_idlanguage INT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS forumcategory_language_idlanguage ON forumcategory (language_idlanguage);

ALTER TABLE forumtopic
    ADD COLUMN IF NOT EXISTS language_idlanguage INT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS forumtopic_language_idlanguage ON forumtopic (language_idlanguage);

-- Update schema version
UPDATE schema_version SET version = 54 WHERE version = 53;
