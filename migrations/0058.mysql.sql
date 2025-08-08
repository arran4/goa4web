-- Allow null language for forum categories and topics
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumcategory MODIFY COLUMN language_idlanguage INT NULL;

UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumtopic MODIFY COLUMN language_idlanguage INT NULL;

-- Update schema version
UPDATE schema_version SET version = 58 WHERE version = 57;
