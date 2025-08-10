-- BLOGS
ALTER TABLE blogs
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- COMMENTS
ALTER TABLE comments
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- FAQ
ALTER TABLE faq
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- LINKER
ALTER TABLE linker
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- LINKER_QUEUE
ALTER TABLE linker_queue
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- PREFERENCES
ALTER TABLE preferences
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- SITE_NEWS
ALTER TABLE site_news
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- WRITING
ALTER TABLE writing
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- DEACTIVATED_COMMENTS
ALTER TABLE deactivated_comments
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- DEACTIVATED_WRITINGS
ALTER TABLE deactivated_writings
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- DEACTIVATED_BLOGS
ALTER TABLE deactivated_blogs
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- DEACTIVATED_LINKER
ALTER TABLE deactivated_linker
  MODIFY COLUMN language_idlanguage INT UNSIGNED NULL,
  ALTER COLUMN language_idlanguage DROP DEFAULT;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;

-- forumthread_id in blogs
ALTER TABLE blogs
  MODIFY COLUMN forumthread_id INT UNSIGNED NULL,
  ALTER COLUMN forumthread_id DROP DEFAULT;
UPDATE blogs SET forumthread_id = NULL WHERE forumthread_id = 0;

-- schema version
UPDATE schema_version SET version = 60 WHERE version = 59;
