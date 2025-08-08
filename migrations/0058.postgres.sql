-- Allow null language for forum categories and topics
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP DEFAULT;

-- Allow FAQ categories to be nullable
ALTER TABLE faq
    MODIFY COLUMN faqCategories_idfaqCategories INT NULL DEFAULT NULL;

UPDATE faq SET faqCategories_idfaqCategories = NULL WHERE faqCategories_idfaqCategories = 0;

-- Allow NULL imageboard references
ALTER TABLE imageboard ALTER COLUMN imageboard_idimageboard DROP NOT NULL;
ALTER TABLE imageboard ALTER COLUMN imageboard_idimageboard DROP DEFAULT;
ALTER TABLE imagepost ALTER COLUMN imageboard_idimageboard DROP NOT NULL;
ALTER TABLE imagepost ALTER COLUMN imageboard_idimageboard DROP DEFAULT;
ALTER TABLE deactivated_imageposts ALTER COLUMN imageboard_idimageboard DROP NOT NULL;
ALTER TABLE deactivated_imageposts ALTER COLUMN imageboard_idimageboard DROP DEFAULT;

UPDATE imageboard SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE imagepost SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE deactivated_imageposts SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;

-- Update schema version
UPDATE schema_version SET version = 58 WHERE version = 57;
