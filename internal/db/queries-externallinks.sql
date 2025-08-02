-- name: RegisterExternalLinkClick :exec
INSERT INTO external_links (url, clicks)
VALUES (?, 1)
ON DUPLICATE KEY UPDATE clicks = clicks + 1;

-- name: GetExternalLink :one
SELECT * FROM external_links WHERE url = ? LIMIT 1;
-- name: ListExternalLinks :many
SELECT * FROM external_links
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
-- name: DeleteExternalLink :exec
DELETE FROM external_links
WHERE external_links.id = sqlc.arg(id)
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id AND r.is_admin = 1
    WHERE ur.users_idusers = sqlc.arg(admin_id)
  );

-- name: ClearExternalLinkCache :exec
UPDATE external_links
SET card_image_cache = NULL,
    favicon_cache    = NULL,
    updated_at       = CURRENT_TIMESTAMP,
    updated_by       = sqlc.arg(updated_by)
WHERE external_links.id = sqlc.arg(id)
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id AND r.is_admin = 1
    WHERE ur.users_idusers = sqlc.arg(admin_id)
  );
