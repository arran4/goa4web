-- Allow NULL language references
ALTER TABLE blogs ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE blogs ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE comments ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE comments ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE faq ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE faq ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE forumcategory ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE forumcategory SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE forumtopic ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE forumtopic SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE linker ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE linker_queue ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE linker_queue ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE preferences ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE preferences ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE site_news ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE site_news ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE user_language ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE user_language ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE user_language SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE writing ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE writing ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_comments ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE deactivated_comments ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_writings ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE deactivated_writings ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_blogs ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE deactivated_blogs ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

ALTER TABLE deactivated_linker ALTER COLUMN language_idlanguage DROP DEFAULT;
ALTER TABLE deactivated_linker ALTER COLUMN language_idlanguage DROP NOT NULL;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

UPDATE schema_version SET version = 58 WHERE version = 57;
