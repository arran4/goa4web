-- Add deactivated user and comment tables; mark deletions with timestamps
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

ALTER TABLE comments
    ADD COLUMN IF NOT EXISTS deleted_at DATETIME DEFAULT NULL;

CREATE TABLE IF NOT EXISTS deactivated_users (
    idusers INT NOT NULL,
    email TINYTEXT,
    passwd TINYTEXT,
    passwd_algorithm TINYTEXT,
    username TINYTEXT,
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
    text LONGTEXT,
    deleted_at DATETIME DEFAULT NULL,
    restored_at DATETIME DEFAULT NULL,
    PRIMARY KEY (idcomments)
);

-- Record upgrade to schema version 8
UPDATE schema_version SET version = 8 WHERE version = 7;
