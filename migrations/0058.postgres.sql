-- Allow FAQ categories to be nullable
ALTER TABLE faq
    MODIFY COLUMN faqCategories_idfaqCategories INT NULL DEFAULT NULL;

UPDATE faq SET faqCategories_idfaqCategories = NULL WHERE faqCategories_idfaqCategories = 0;

UPDATE schema_version SET version = 58 WHERE version = 57;
