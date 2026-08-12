-- +goose Up
CREATE TABLE user_passkeys (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    credential_id BLOB NOT NULL,
    public_key BLOB NOT NULL,
    attestation_type VARCHAR(255) NOT NULL,
    aaguid BLOB NOT NULL,
    sign_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id),
    KEY user_passkeys_user_idx (user_id),
    UNIQUE KEY user_passkeys_cred_idx (credential_id(255))
);
-- +goose Down
DROP TABLE user_passkeys;
