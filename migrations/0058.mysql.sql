-- Allow null language for forum categories and topics
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumcategory MODIFY COLUMN language_idlanguage INT NULL;

UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE forumtopic MODIFY COLUMN language_idlanguage INT NULL;

-- Allow FAQ categories to be nullable
ALTER TABLE faq
    MODIFY COLUMN faqCategories_idfaqCategories INT NULL DEFAULT NULL;

UPDATE faq SET faqCategories_idfaqCategories = NULL WHERE faqCategories_idfaqCategories = 0;

-- Allow NULL imageboard references
ALTER TABLE imageboard MODIFY imageboard_idimageboard INT NULL;
ALTER TABLE imagepost MODIFY imageboard_idimageboard INT NULL;
ALTER TABLE deactivated_imageposts MODIFY imageboard_idimageboard INT NULL;

UPDATE imageboard SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE imagepost SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;
UPDATE deactivated_imageposts SET imageboard_idimageboard = NULL WHERE imageboard_idimageboard = 0;

-- Update schema version
UPDATE schema_version SET version = 58 WHERE version = 57;
