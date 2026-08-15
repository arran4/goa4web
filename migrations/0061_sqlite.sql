-- +goose Up
-- ALTER TABLE faq CHANGE COLUMN idfaq id INT NOT NULL AUTO_INCREMENT;

-- ALTER TABLE faq CHANGE COLUMN faqCategories_idfaqCategories faq_category_id INT NULL DEFAULT NULL;

-- ALTER TABLE faq_categories CHANGE COLUMN idfaqCategories id INT NOT NULL AUTO_INCREMENT;