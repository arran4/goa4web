-- +goose Up
ALTER TABLE faq RENAME COLUMN idfaq TO id;
ALTER TABLE faq RENAME COLUMN faqCategories_idfaqCategories TO faq_category_id;
ALTER TABLE faq_categories RENAME COLUMN idfaqCategories TO id;
UPDATE schema_version SET version = 61;