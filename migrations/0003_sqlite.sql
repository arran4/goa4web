-- +goose Up
CREATE TABLE IF NOT EXISTS sessions (
session_id TEXT NOT NULL,
users_idusers int NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (session_id)
);