ALTER TABLE blogs MODIFY COLUMN language_idlanguage INT NULL;
UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE comments MODIFY COLUMN language_idlanguage INT NULL;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE faq MODIFY COLUMN language_idlanguage INT NULL;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker MODIFY COLUMN language_idlanguage INT NULL;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker_queue MODIFY COLUMN language_idlanguage INT NULL;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE preferences MODIFY COLUMN language_idlanguage INT NULL;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE site_news MODIFY COLUMN language_idlanguage INT NULL;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE writing MODIFY COLUMN language_idlanguage INT NULL;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_comments MODIFY COLUMN language_idlanguage INT NULL;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_writings MODIFY COLUMN language_idlanguage INT NULL;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_blogs MODIFY COLUMN language_idlanguage INT NULL;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_linker MODIFY COLUMN language_idlanguage INT NULL;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

UPDATE blogs SET forumthread_id = NULL WHERE forumthread_id = 0;

ALTER TABLE blogs
    MODIFY COLUMN forumthread_id int(10) DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 60 WHERE version = 59;
