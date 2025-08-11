ALTER TABLE faq
   CHANGE COLUMN faq_category_id category_id INT NULL DEFAULT NULL;

ALTER TABLE faq_categories
   ADD COLUMN parent_category_id INT NULL DEFAULT NULL,
   ADD COLUMN language_id INT NULL DEFAULT NULL;

ALTER TABLE language CHANGE COLUMN idlanguage id INT NOT NULL AUTO_INCREMENT;
ALTER TABLE blogs CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE comments CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE faq CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE forumcategory CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE forumtopic CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE linker CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE linker_queue CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE preferences CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE site_news CHANGE COLUMN language_idlanguage language_id INT DEFAULT NULL;
ALTER TABLE forumcategory CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE forumtopic CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE linker CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE linker_queue CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE preferences CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE site_news CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE writing CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE deactivated_comments CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE deactivated_writings CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE deactivated_blogs CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE deactivated_linker CHANGE COLUMN language_idlanguage language_id INT NULL;
ALTER TABLE user_language CHANGE COLUMN language_idlanguage language_id INT NOT NULL DEFAULT 0;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
