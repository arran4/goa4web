-- Add word_count columns to search tables
ALTER TABLE comments_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;
ALTER TABLE site_news_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;
ALTER TABLE blogs_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;
ALTER TABLE linker_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;
ALTER TABLE writing_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;
ALTER TABLE imagepost_search ADD COLUMN word_count INT NOT NULL DEFAULT 1;

-- Update schema version
UPDATE schema_version SET version = 48 WHERE version = 47;
