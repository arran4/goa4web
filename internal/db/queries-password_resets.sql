-- name: CreatePasswordResetForUser :exec
INSERT INTO pending_passwords (user_id, passwd, passwd_algorithm, verification_code)
VALUES (?, ?, ?, ?);

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

-- name: AdminGetPasswordResetByID :one
SELECT id, user_id, passwd, passwd_algorithm, verification_code, created_at, verified_at
FROM pending_passwords
WHERE id = ?;

-- name: AdminListPasswordResets :many
SELECT pp.id, pp.user_id, u.username, pp.created_at, pp.verified_at
FROM pending_passwords pp
JOIN users u ON pp.user_id = u.idusers
WHERE
    (
        (sqlc.narg('status') = 'pending' AND pp.verified_at IS NULL) OR
        (sqlc.narg('status') = 'verified' AND pp.verified_at IS NOT NULL) OR
        (sqlc.narg('status') IS NULL)
    )
    AND (sqlc.narg('user_id') IS NULL OR pp.user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('created_before') IS NULL OR pp.created_at < sqlc.narg('created_before'))
ORDER BY pp.created_at DESC
LIMIT ? OFFSET ?;

-- name: AdminCountPasswordResets :one
SELECT COUNT(*)
FROM pending_passwords pp
WHERE
    (
        (sqlc.narg('status') = 'pending' AND pp.verified_at IS NULL) OR
        (sqlc.narg('status') = 'verified' AND pp.verified_at IS NOT NULL) OR
        (sqlc.narg('status') IS NULL)
    )
    AND (sqlc.narg('user_id') IS NULL OR pp.user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('created_before') IS NULL OR pp.created_at < sqlc.narg('created_before'));

-- name: AdminCountPendingPasswordResetsByUser :many
SELECT user_id, COUNT(*) as count
FROM pending_passwords
WHERE verified_at IS NULL
GROUP BY user_id;
