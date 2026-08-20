-- name: CreateAPIKey :execlastid
INSERT INTO api_keys (users_idusers, api_key, name, scopes, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE api_key = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);

-- name: ListAPIKeysByUser :many
SELECT * FROM api_keys WHERE users_idusers = ? ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked_at = CURRENT_TIMESTAMP WHERE id = ? AND users_idusers = ?;
