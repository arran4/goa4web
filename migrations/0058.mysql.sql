-- Allow NULL language references
ALTER TABLE blogs MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE comments MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE faq MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE forumcategory MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE forumtopic MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker_queue MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE preferences MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE site_news MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE user_language MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE user_language SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE writing MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_comments MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_writings MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_blogs MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_linker MODIFY language_idlanguage INT NULL DEFAULT NULL;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

UPDATE schema_version SET version = 58 WHERE version = 57;
