CREATE TABLE forumtopic_public_labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forumtopic_idforumtopic INT NOT NULL,
    label TEXT NOT NULL,
    UNIQUE (forumtopic_idforumtopic, label)
);

CREATE TABLE forumtopic_private_labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forumtopic_idforumtopic INT NOT NULL,
    users_idusers INT NOT NULL,
    label TEXT NOT NULL,
    invert INT NOT NULL DEFAULT 0,
    UNIQUE (forumtopic_idforumtopic, users_idusers, label)
);

CREATE TABLE content_label_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item TEXT NOT NULL,
    item_id INT NOT NULL,
    label TEXT NOT NULL,
    UNIQUE (item, item_id, label)
);

-- Update schema version
UPDATE schema_version SET version = 62 WHERE version = 61;
