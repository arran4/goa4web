CREATE TABLE user_emails (
    id int NOT NULL AUTO_INCREMENT,
    users_idusers int NOT NULL,
    email tinytext NOT NULL,
    verified tinyint(1) NOT NULL DEFAULT 0,
    last_verification_code varchar(64) DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY user_emails_unique (users_idusers, email(255)),
    KEY user_emails_user_idx (users_idusers)
);

INSERT INTO user_emails (users_idusers, email)
SELECT idusers, email FROM users WHERE email IS NOT NULL AND email != '';

-- Record upgrade to schema version 26
UPDATE schema_version SET version = 26 WHERE version = 25;
