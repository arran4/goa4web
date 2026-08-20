-- +goose Up
DROP TABLE IF EXISTS writing_user_permissions;

CREATE TABLE IF NOT EXISTS admin_user_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO roles (name) VALUES ('rejected');
UPDATE schema_version SET version = 40;