-- Rename search columns to snake_case
ALTER TABLE blogsSearch CHANGE COLUMN blogs_idblogs blog_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE siteNewsSearch CHANGE COLUMN siteNews_idsiteNews site_news_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 19
UPDATE schema_version SET version = 19 WHERE version = 18;
