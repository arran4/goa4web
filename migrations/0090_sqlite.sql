-- +goose Up
CREATE TABLE IF NOT EXISTS user_passkeys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INT NOT NULL,
    credential_id BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    aaguid BLOB NOT NULL,
    sign_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    UNIQUE (credential_id)
);
UPDATE schema_version SET version = 90;

-- +goose Down
DROP TABLE IF EXISTS user_passkeys;