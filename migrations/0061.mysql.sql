ALTER TABLE faq CHANGE COLUMN idfaq id INT NOT NULL AUTO_INCREMENT;
ALTER TABLE faq CHANGE COLUMN faqCategories_idfaqCategories faq_category_id INT NULL DEFAULT NULL;
ALTER TABLE faq_categories CHANGE COLUMN idfaqCategories id INT NOT NULL AUTO_INCREMENT;

-- Update schema version
UPDATE schema_version SET version = 61 WHERE version = 60;
