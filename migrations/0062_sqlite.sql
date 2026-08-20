-- +goose Up
CREATE TABLE IF NOT EXISTS forumtopic_public_labels (
id INTEGER PRIMARY KEY AUTOINCREMENT,
forumtopic_idforumtopic INT NOT NULL,
label TEXT NOT NULL,
UNIQUE (forumtopic_idforumtopic, label)
);

CREATE TABLE IF NOT EXISTS forumtopic_private_labels (
id INTEGER PRIMARY KEY AUTOINCREMENT,
forumtopic_idforumtopic INT NOT NULL,
users_idusers INT NOT NULL,
label TEXT NOT NULL,
invert INTEGER NOT NULL DEFAULT 0,
UNIQUE (forumtopic_idforumtopic, users_idusers, label)
);

CREATE TABLE IF NOT EXISTS content_label_status (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id INT NOT NULL,
label TEXT NOT NULL,
UNIQUE (item, item_id, label)
);
UPDATE schema_version SET version = 62;