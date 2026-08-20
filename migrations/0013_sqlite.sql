-- +goose Up
CREATE TABLE IF NOT EXISTS passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
UPDATE schema_version SET version = 13;