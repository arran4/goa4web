-- Rename tables to snake_case
RENAME TABLE siteNews TO site_news;
RENAME TABLE siteNewsSearch TO site_news_search;
RENAME TABLE blogsSearch TO blogs_search;
RENAME TABLE commentsSearch TO comments_search;
RENAME TABLE imagepostSearch TO imagepost_search;
RENAME TABLE linkerCategory TO linker_category;
RENAME TABLE linkerQueue TO linker_queue;
RENAME TABLE linkerSearch TO linker_search;
RENAME TABLE writingSearch TO writing_search;
RENAME TABLE writingApprovedUsers TO writing_approved_users;
RENAME TABLE faqCategories TO faq_categories;

-- Record upgrade to schema version 24
UPDATE schema_version SET version = 24 WHERE version = 23;
