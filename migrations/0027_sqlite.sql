-- +goose Up
-- ALTER TABLE user_emails
    CHANGE COLUMN users_idusers user_id int NOT NULL,
    DROP COLUMN verified,
    ADD COLUMN verified_at datetime DEFAULT NULL,
    DROP INDEX user_emails_unique,
    ADD UNIQUE KEY user_emails_email_idx (email(255));

INSERT INTO user_emails (user_id, email, verified_at)
SELECT idusers, email, NOW() FROM users
WHERE email IS NOT NULL AND email != '' AND idusers NOT IN (SELECT user_id FROM user_emails);

-- ALTER TABLE users
    DROP INDEX users_email_idx,
    DROP COLUMN email;

CREATE TABLE pending_passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
verification_code TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
verified_at DATETIME DEFAULT NULL,
UNIQUE (verification_code)
);