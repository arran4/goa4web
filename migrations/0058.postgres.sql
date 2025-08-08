-- Allow null language for forum categories and topics
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP DEFAULT;

-- Update schema version
UPDATE schema_version SET version = 58 WHERE version = 57;
