ALTER TABLE user_emails
    CHANGE COLUMN users_idusers user_id int NOT NULL,
    DROP COLUMN verified,
    ADD COLUMN verified_at datetime DEFAULT NULL,
    DROP INDEX user_emails_unique,
    ADD UNIQUE KEY user_emails_email_idx (email(255)),
    ADD KEY user_emails_user_idx (user_id);

INSERT INTO user_emails (user_id, email, verified_at)
SELECT idusers, email, NOW() FROM users
WHERE email IS NOT NULL AND email != '' AND idusers NOT IN (SELECT user_id FROM user_emails);

ALTER TABLE users
    DROP INDEX users_email_idx,
    DROP COLUMN email;

CREATE TABLE pending_passwords (
    id int NOT NULL AUTO_INCREMENT,
    user_id int NOT NULL,
    passwd tinytext NOT NULL,
    passwd_algorithm tinytext,
    verification_code varchar(64) NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at datetime DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY pending_password_code_idx (verification_code),
    KEY pending_password_user_idx (user_id)
);

-- Record upgrade to schema version 27
UPDATE schema_version SET version = 27 WHERE version = 26;
