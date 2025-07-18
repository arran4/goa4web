-- name: CreatePasswordReset :exec
INSERT INTO pending_passwords (user_id, passwd, passwd_algorithm, verification_code)
VALUES (?, ?, ?, ?);

-- name: GetPasswordResetByUser :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE user_id = ? AND verified_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPasswordResetByCode :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE verification_code = ? AND verified_at IS NULL;

-- name: MarkPasswordResetVerified :exec
UPDATE pending_passwords SET verified_at = NOW() WHERE id = ?;

-- name: DeletePasswordReset :exec
DELETE FROM pending_passwords WHERE id = ?;
