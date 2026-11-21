-- +goose Up
-- +goose StatementBegin
-- Keep the historical envelope columns while allowing MIME-based queue inserts.
CREATE TABLE pending_emails_0105 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_email TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME DEFAULT NULL,
    html_body TEXT,
    error_count INT NOT NULL DEFAULT 0,
    to_user_id INT DEFAULT NULL,
    direct_email INTEGER NOT NULL DEFAULT 0,
    digest_delivery_id INT DEFAULT NULL
);

INSERT INTO pending_emails_0105
    (id, to_email, subject, body, created_at, sent_at, html_body, error_count, to_user_id, direct_email, digest_delivery_id)
SELECT id, to_email, subject, body, created_at, sent_at, html_body, error_count, NULLIF(to_user_id, 0), direct_email, digest_delivery_id
FROM pending_emails;

INSERT INTO sqlite_sequence (name, seq)
SELECT 'pending_emails_0105', seq FROM sqlite_sequence
WHERE name = 'pending_emails' AND NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = 'pending_emails_0105');
UPDATE sqlite_sequence
SET seq = MAX(seq, COALESCE((SELECT seq FROM sqlite_sequence WHERE name = 'pending_emails'), 0))
WHERE name = 'pending_emails_0105';

DROP TABLE pending_emails;
ALTER TABLE pending_emails_0105 RENAME TO pending_emails;
CREATE UNIQUE INDEX pending_emails_digest_delivery_idx ON pending_emails (digest_delivery_id);
UPDATE schema_version SET version = 105;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE pending_emails_0104 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME DEFAULT NULL,
    html_body TEXT,
    error_count INT NOT NULL DEFAULT 0,
    to_user_id INT NOT NULL DEFAULT 0,
    direct_email INTEGER NOT NULL DEFAULT 0,
    digest_delivery_id INT DEFAULT NULL
);

INSERT INTO pending_emails_0104
    (id, to_email, subject, body, created_at, sent_at, html_body, error_count, to_user_id, direct_email, digest_delivery_id)
SELECT id, to_email, subject, body, created_at, sent_at, html_body, error_count, COALESCE(to_user_id, 0), direct_email, digest_delivery_id
FROM pending_emails;

INSERT INTO sqlite_sequence (name, seq)
SELECT 'pending_emails_0104', seq FROM sqlite_sequence
WHERE name = 'pending_emails' AND NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = 'pending_emails_0104');
UPDATE sqlite_sequence
SET seq = MAX(seq, COALESCE((SELECT seq FROM sqlite_sequence WHERE name = 'pending_emails'), 0))
WHERE name = 'pending_emails_0104';

DROP TABLE pending_emails;
ALTER TABLE pending_emails_0104 RENAME TO pending_emails;
CREATE UNIQUE INDEX pending_emails_digest_delivery_idx ON pending_emails (digest_delivery_id);
UPDATE schema_version SET version = 104;
-- +goose StatementEnd
