-- +goose Up
CREATE TABLE user_emails (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers int NOT NULL,
email TEXT NOT NULL,
verified INTEGER NOT NULL DEFAULT 0,
last_verification_code TEXT DEFAULT NULL,
UNIQUE (users_idusers, email)
);

INSERT INTO user_emails (users_idusers, email)
SELECT idusers, email FROM users WHERE email IS NOT NULL AND email != '';