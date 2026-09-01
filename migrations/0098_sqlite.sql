-- +goose Up
-- Repair the nullable-language schema drift left by migration 0060. These
-- definitions preserve every column subsequently added to the affected tables.

CREATE TABLE blogs_0098 (
    idblogs INTEGER PRIMARY KEY AUTOINCREMENT,
    forumthread_id INTEGER DEFAULT NULL,
    users_idusers INTEGER NOT NULL DEFAULT 0,
    language_id INTEGER DEFAULT NULL,
    blog TEXT DEFAULT NULL,
    written DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    last_index DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO blogs_0098 SELECT * FROM blogs;
DROP TABLE blogs;
ALTER TABLE blogs_0098 RENAME TO blogs;

CREATE TABLE comments_0098 (
    idcomments INTEGER PRIMARY KEY AUTOINCREMENT,
    forumthread_id INTEGER NOT NULL DEFAULT 0,
    users_idusers INTEGER NOT NULL DEFAULT 0,
    language_id INTEGER DEFAULT NULL,
    written DATETIME DEFAULT NULL,
    text TEXT DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    last_index DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO comments_0098 SELECT * FROM comments;
DROP TABLE comments;
ALTER TABLE comments_0098 RENAME TO comments;

CREATE TABLE faq_0098 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL DEFAULT 0,
    language_id INTEGER DEFAULT NULL,
    author_id INTEGER NOT NULL DEFAULT 0,
    answer TEXT DEFAULT NULL,
    question TEXT DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    priority INT NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT NULL,
    description TEXT DEFAULT ''
);
INSERT INTO faq_0098 SELECT * FROM faq;
DROP TABLE faq;
ALTER TABLE faq_0098 RENAME TO faq;
CREATE INDEX faq_priority_idx ON faq (priority);

CREATE TABLE linker_0098 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    language_id INTEGER DEFAULT NULL,
    author_id INTEGER NOT NULL DEFAULT 0,
    category_id INTEGER NOT NULL DEFAULT 0,
    thread_id INTEGER NOT NULL DEFAULT 0,
    title TEXT DEFAULT NULL,
    url TEXT DEFAULT NULL,
    description TEXT DEFAULT NULL,
    listed DATETIME DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    last_index DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO linker_0098 SELECT * FROM linker;
DROP TABLE linker;
ALTER TABLE linker_0098 RENAME TO linker;

CREATE TABLE linker_queue_0098 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    language_id INTEGER DEFAULT NULL,
    submitter_id INTEGER NOT NULL DEFAULT 0,
    category_id INTEGER NOT NULL DEFAULT 0,
    title TEXT DEFAULT NULL,
    url TEXT DEFAULT NULL,
    description TEXT DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO linker_queue_0098 SELECT * FROM linker_queue;
DROP TABLE linker_queue;
ALTER TABLE linker_queue_0098 RENAME TO linker_queue;

CREATE TABLE preferences_0098 (
    idpreferences INTEGER PRIMARY KEY AUTOINCREMENT,
    language_id INTEGER DEFAULT NULL,
    users_idusers INTEGER NOT NULL DEFAULT 0,
    emailforumupdates INTEGER DEFAULT 0,
    page_size INT NOT NULL DEFAULT 15,
    auto_subscribe_replies TINYINT(1) NOT NULL DEFAULT 1,
    timezone TEXT DEFAULT NULL,
    custom_css TEXT,
    daily_digest_hour INT DEFAULT NULL,
    daily_digest_mark_read TINYINT(1) NOT NULL DEFAULT 0,
    last_digest_sent_at DATETIME DEFAULT NULL,
    weekly_digest_day INT DEFAULT NULL,
    weekly_digest_hour INT DEFAULT NULL,
    last_weekly_digest_sent_at DATETIME DEFAULT NULL,
    monthly_digest_day INT DEFAULT NULL,
    monthly_digest_hour INT DEFAULT NULL,
    last_monthly_digest_sent_at DATETIME DEFAULT NULL,
    image_safe_dimension TEXT
);
INSERT INTO preferences_0098 SELECT * FROM preferences;
DROP TABLE preferences;
ALTER TABLE preferences_0098 RENAME TO preferences;

CREATE TABLE site_news_0098 (
    idsiteNews INTEGER PRIMARY KEY AUTOINCREMENT,
    forumthread_id INTEGER NOT NULL DEFAULT 0,
    language_id INTEGER DEFAULT NULL,
    users_idusers INTEGER NOT NULL DEFAULT 0,
    news TEXT DEFAULT NULL,
    occurred DATETIME DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    last_index DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO site_news_0098 SELECT * FROM site_news;
DROP TABLE site_news;
ALTER TABLE site_news_0098 RENAME TO site_news;

CREATE TABLE writing_0098 (
    idwriting INTEGER PRIMARY KEY AUTOINCREMENT,
    users_idusers INTEGER NOT NULL DEFAULT 0,
    forumthread_id INTEGER NOT NULL DEFAULT 0,
    language_id INTEGER DEFAULT NULL,
    writing_category_id INTEGER NOT NULL DEFAULT 0,
    title TEXT DEFAULT NULL,
    published DATETIME DEFAULT NULL,
    writing TEXT DEFAULT NULL,
    abstract TEXT DEFAULT NULL,
    private INTEGER DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    last_index DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL
);
INSERT INTO writing_0098 SELECT * FROM writing;
DROP TABLE writing;
ALTER TABLE writing_0098 RENAME TO writing;

CREATE TABLE deactivated_comments_0098 (
    idcomments INT NOT NULL,
    forumthread_id INT NOT NULL,
    users_idusers INT NOT NULL,
    language_id INT DEFAULT NULL,
    written DATETIME,
    text TEXT,
    deleted_at DATETIME DEFAULT NULL,
    restored_at DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL,
    PRIMARY KEY (idcomments)
);
INSERT INTO deactivated_comments_0098 SELECT * FROM deactivated_comments;
DROP TABLE deactivated_comments;
ALTER TABLE deactivated_comments_0098 RENAME TO deactivated_comments;

CREATE TABLE deactivated_writings_0098 (
    idwriting INT NOT NULL,
    users_idusers INT NOT NULL,
    forumthread_id INT NOT NULL,
    language_id INT DEFAULT NULL,
    writingCategory_idwritingCategory INT NOT NULL,
    title TEXT,
    published DATETIME,
    writing TEXT,
    abstract TEXT,
    private INTEGER DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    restored_at DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL,
    PRIMARY KEY (idwriting)
);
INSERT INTO deactivated_writings_0098 SELECT * FROM deactivated_writings;
DROP TABLE deactivated_writings;
ALTER TABLE deactivated_writings_0098 RENAME TO deactivated_writings;

CREATE TABLE deactivated_blogs_0098 (
    idblogs INT NOT NULL,
    forumthread_id INT NOT NULL,
    users_idusers INT NOT NULL,
    language_id INT DEFAULT NULL,
    blog TEXT,
    written DATETIME,
    deleted_at DATETIME DEFAULT NULL,
    restored_at DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL,
    PRIMARY KEY (idblogs)
);
INSERT INTO deactivated_blogs_0098 SELECT * FROM deactivated_blogs;
DROP TABLE deactivated_blogs;
ALTER TABLE deactivated_blogs_0098 RENAME TO deactivated_blogs;

CREATE TABLE deactivated_linker_0098 (
    id INT NOT NULL,
    language_id INT DEFAULT NULL,
    author_id INT NOT NULL,
    category_id INT NOT NULL,
    thread_id INT NOT NULL,
    title TEXT,
    url TEXT,
    description TEXT,
    listed DATETIME,
    deleted_at DATETIME DEFAULT NULL,
    restored_at DATETIME DEFAULT NULL,
    timezone TEXT DEFAULT NULL,
    PRIMARY KEY (id)
);
INSERT INTO deactivated_linker_0098 SELECT * FROM deactivated_linker;
DROP TABLE deactivated_linker;
ALTER TABLE deactivated_linker_0098 RENAME TO deactivated_linker;

UPDATE schema_version SET version = 98;

-- +goose Down
UPDATE schema_version SET version = 97;
