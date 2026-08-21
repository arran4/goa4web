-- +goose Up
ALTER TABLE users ADD COLUMN deleted_at DATETIME DEFAULT NULL;
ALTER TABLE comments ADD COLUMN deleted_at DATETIME DEFAULT NULL;

CREATE TABLE IF NOT EXISTS deactivated_users (
idusers INT NOT NULL,
email TEXT,
passwd TEXT,
passwd_algorithm TEXT,
username TEXT,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idusers)
);

CREATE TABLE IF NOT EXISTS deactivated_comments (
idcomments INT NOT NULL,
forumthread_idforumthread INT NOT NULL,
users_idusers INT NOT NULL,
language_idlanguage INT NOT NULL,
written DATETIME,
TEXT TEXT,
deleted_at DATETIME DEFAULT NULL,
restored_at DATETIME DEFAULT NULL,
PRIMARY KEY (idcomments)
);
UPDATE schema_version SET version = 8;