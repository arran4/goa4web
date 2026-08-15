-- +goose Up
DROP TABLE IF EXISTS writing_user_permissions;

CREATE TABLE admin_user_comments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
comment TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (name) VALUES ('rejected') ON DUPLICATE KEY UPDATE name=name;