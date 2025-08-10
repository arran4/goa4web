UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE blogs ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE blogs ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE comments ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE comments ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE faq ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE faq ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE linker ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE linker ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE linker_queue ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE linker_queue ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE preferences ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE preferences ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE site_news ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE site_news ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE writing ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE writing ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE deactivated_comments ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE deactivated_comments ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE deactivated_writings ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE deactivated_writings ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE deactivated_blogs ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE deactivated_blogs ALTER COLUMN language_idlanguage DROP DEFAULT;

UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;
ALTER TABLE deactivated_linker ALTER COLUMN language_idlanguage DROP NOT NULL;
ALTER TABLE deactivated_linker ALTER COLUMN language_idlanguage DROP DEFAULT;

-- Update schema version
UPDATE schema_version SET version = 60 WHERE version = 59;
