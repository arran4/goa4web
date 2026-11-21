-- name: CreateAPIKey :execlastid
INSERT INTO api_keys (users_idusers, api_key, name, scopes, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: SystemGetActiveAPIKeyByHash :one
SELECT k.*
FROM api_keys k
JOIN users u ON u.idusers = k.users_idusers
WHERE k.api_key = ?
  AND k.revoked_at IS NULL
  AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)
  AND u.deleted_at IS NULL;

-- name: ListAPIKeysByUser :many
SELECT * FROM api_keys WHERE users_idusers = ? ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ? AND users_idusers = ?;
