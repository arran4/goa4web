-- +goose Up
ALTER TABLE siteNews RENAME TO site_news;
ALTER TABLE siteNewsSearch RENAME TO site_news_search;
ALTER TABLE blogsSearch RENAME TO blogs_search;
ALTER TABLE commentsSearch RENAME TO comments_search;
ALTER TABLE imagepostSearch RENAME TO imagepost_search;
ALTER TABLE linkerCategory RENAME TO linker_category;
ALTER TABLE linkerQueue RENAME TO linker_queue;
ALTER TABLE linkerSearch RENAME TO linker_search;
ALTER TABLE writingSearch RENAME TO writing_search;
ALTER TABLE writingApprovedUsers RENAME TO writing_approved_users;
ALTER TABLE faqCategories RENAME TO faq_categories;
UPDATE schema_version SET version = 24;