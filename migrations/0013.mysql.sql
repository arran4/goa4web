-- Split password into separate table
CREATE TABLE IF NOT EXISTS passwords (
    id INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    passwd TINYTEXT NOT NULL,
    passwd_algorithm TINYTEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY passwords_user_idx (users_idusers)
);

ALTER TABLE users
    DROP COLUMN IF EXISTS passwd,
    DROP COLUMN IF EXISTS passwd_algorithm;

-- Record upgrade to schema version 13
UPDATE schema_version SET version = 13 WHERE version = 12;
