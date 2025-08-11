ALTER TABLE faq
    CHANGE COLUMN faq_category_id category_id INT NULL DEFAULT NULL,
    CHANGE COLUMN language_idlanguage language_id INT NULL DEFAULT NULL,
    CHANGE COLUMN users_idusers author_id INT NOT NULL DEFAULT 0;

ALTER TABLE faq_categories
    ADD COLUMN parent_category_id INT NULL DEFAULT NULL,
    ADD COLUMN language_id INT NULL DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
