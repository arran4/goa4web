CREATE TABLE IF NOT EXISTS uploaded_images (
    iduploadedimage INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    path TINYTEXT,
    thumbnail TINYTEXT,
    file_size INT NOT NULL,
    uploaded DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (iduploadedimage),
    KEY uploaded_images_user_idx (users_idusers)
);

-- Rename search columns to snake_case
ALTER TABLE blogsSearch CHANGE COLUMN blogs_idblogs blog_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE siteNewsSearch CHANGE COLUMN siteNews_idsiteNews site_news_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 19
UPDATE schema_version SET version = 19 WHERE version = 18;
