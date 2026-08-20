-- +goose Up
CREATE TABLE IF NOT EXISTS thread_images (
idthread_image INTEGER PRIMARY KEY AUTOINCREMENT,
forumthread_id INT NOT NULL,
path TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
UPDATE schema_version SET version = 77;