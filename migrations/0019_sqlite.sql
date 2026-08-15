-- +goose Up
CREATE TABLE IF NOT EXISTS uploaded_images (
iduploadedimage INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
path TEXT,
thumbnail TEXT,
file_size INT NOT NULL,
uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ALTER TABLE blogsSearch CHANGE COLUMN blogs_idblogs blog_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE siteNewsSearch CHANGE COLUMN siteNews_idsiteNews site_news_id int(10) NOT NULL DEFAULT 0;