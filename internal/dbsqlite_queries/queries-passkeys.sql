-- name: InsertPasskey :exec
INSERT INTO user_passkeys (
    user_id, name, backup_eligible, backup_state, credential_id, public_key, attestation_type, aaguid, sign_count
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetPasskeysByUserID :many
SELECT * FROM user_passkeys WHERE user_id = ?;

-- name: GetPasskeyByCredentialID :one
SELECT * FROM user_passkeys WHERE credential_id = ?;

-- name: UpdatePasskeyAfterLogin :exec
UPDATE user_passkeys
SET sign_count = ?,
    backup_eligible = COALESCE(backup_eligible, ?),
    backup_state = ?,
    last_used_at = CURRENT_TIMESTAMP
WHERE credential_id = ?;

-- name: DeletePasskey :exec
DELETE FROM user_passkeys WHERE id = ? AND user_id = ?;
