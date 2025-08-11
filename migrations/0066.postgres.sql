ALTER TABLE faq RENAME COLUMN faq_category_id TO category_id;
ALTER TABLE faq RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE faq RENAME COLUMN users_idusers TO author_id;

ALTER TABLE faq_categories ADD COLUMN parent_category_id INT;
ALTER TABLE faq_categories ADD COLUMN language_id INT;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
