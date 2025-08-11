CREATE TABLE forumtopic_public_labels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    forumtopic_idforumtopic INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    UNIQUE KEY forumtopic_public_labels_uq (forumtopic_idforumtopic, label)
);

CREATE TABLE forumtopic_private_labels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    forumtopic_idforumtopic INT NOT NULL,
    users_idusers INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    invert TINYINT(1) NOT NULL DEFAULT 0,
    UNIQUE KEY forumtopic_private_labels_uq (forumtopic_idforumtopic, users_idusers, label)
);

CREATE TABLE content_label_status (
    id INT AUTO_INCREMENT PRIMARY KEY,
    item VARCHAR(255) NOT NULL,
    item_id INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    UNIQUE KEY content_label_status_uq (item, item_id, label)
);

-- Update schema version
UPDATE schema_version SET version = 62 WHERE version = 61;
