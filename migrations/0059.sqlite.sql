-- Allow NULL language references
ALTER TABLE blogs MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE comments MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE faq MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE forumcategory MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE forumtopic MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE linker MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE linker_queue MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE preferences MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE site_news MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE user_language MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE writing MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE deactivated_comments MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE deactivated_writings MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE deactivated_blogs MODIFY language_idlanguage INT NULL DEFAULT NULL;
ALTER TABLE deactivated_linker MODIFY language_idlanguage INT NULL DEFAULT NULL;

UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE user_language SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- Update schema version
UPDATE schema_version SET version = 59 WHERE version = 58;
