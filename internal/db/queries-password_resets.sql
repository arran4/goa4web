-- name: CreatePasswordResetForUser :exec
INSERT INTO pending_passwords (user_id, passwd, passwd_algorithm, verification_code)
VALUES (?, ?, ?, ?);

-- name: CreatePasswordResetTokenForUser :exec
INSERT INTO pending_passwords (user_id, verification_code, passwd, passwd_algorithm)
VALUES (?, ?, NULL, NULL);

-- name: GetPasswordResetByUser :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE user_id = ? AND verified_at IS NULL AND created_at > ?
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPasswordResetByCode :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE verification_code = ? AND verified_at IS NULL AND created_at > ?;

-- name: SystemMarkPasswordResetVerified :exec
UPDATE pending_passwords SET verified_at = NOW() WHERE id = ?;

-- name: SystemDeletePasswordReset :exec
DELETE FROM pending_passwords WHERE id = ?;

-- name: SystemDeletePasswordResetsByUser :execresult
-- Delete all password reset entries for the given user and return the result
DELETE FROM pending_passwords WHERE user_id = ?;

-- name: SystemPurgePasswordResetsBefore :execresult
-- Remove password reset entries that have expired or were already verified
DELETE FROM pending_passwords
WHERE created_at < ? OR verified_at IS NOT NULL;

