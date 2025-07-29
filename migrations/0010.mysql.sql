-- Add deleted_at columns for soft deletion
ALTER TABLE imageboard
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE forumcategory
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE forumtopic
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE forumthread
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE siteNews
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE faq
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;
ALTER TABLE faqCategories
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

-- Record upgrade to schema version 10
UPDATE schema_version SET version = 10 WHERE version = 9;
