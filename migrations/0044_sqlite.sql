-- +goose Up
CREATE TABLE IF NOT EXISTS admin_request_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
request_id INT NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
UPDATE schema_version SET version = 44;