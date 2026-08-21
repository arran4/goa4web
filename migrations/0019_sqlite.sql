-- +goose Up
CREATE TABLE IF NOT EXISTS uploaded_images (
iduploadedimage INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
path TEXT,
thumbnail TEXT,
file_size INT NOT NULL,
uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE blogsSearch RENAME COLUMN blogs_idblogs TO blog_id;
ALTER TABLE siteNewsSearch RENAME COLUMN siteNews_idsiteNews TO site_news_id;
UPDATE schema_version SET version = 19;