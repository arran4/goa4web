-- +goose Up
ALTER TABLE user_emails RENAME COLUMN users_idusers TO user_id;
ALTER TABLE user_emails ADD COLUMN verified_at datetime DEFAULT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS user_emails_email_idx ON user_emails (email);

INSERT INTO user_emails (user_id, email, verified_at)
SELECT idusers, email, CURRENT_TIMESTAMP FROM users
WHERE email IS NOT NULL AND email != '' AND idusers NOT IN (SELECT user_id FROM user_emails);

CREATE TABLE IF NOT EXISTS pending_passwords (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id int NOT NULL,
passwd TEXT NOT NULL,
passwd_algorithm TEXT,
verification_code TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
verified_at DATETIME DEFAULT NULL,
UNIQUE (verification_code)
);
UPDATE schema_version SET version = 27;