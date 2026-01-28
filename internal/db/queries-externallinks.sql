-- name: SystemRegisterExternalLinkClick :exec
INSERT INTO external_links (url, clicks)
VALUES (?, 1)
ON DUPLICATE KEY UPDATE clicks = clicks + 1;

-- name: GetExternalLink :one
SELECT * FROM external_links WHERE url = ? LIMIT 1;

-- name: GetExternalLinkByID :one
SELECT * FROM external_links WHERE id = ? LIMIT 1;

-- name: AdminListExternalLinks :many
SELECT * FROM external_links
ORDER BY created_at DESC
LIMIT ? OFFSET ?;


-- name: AdminDeleteExternalLink :exec
DELETE FROM external_links WHERE id = ?;

-- name: AdminClearExternalLinkCache :exec
UPDATE external_links SET card_image_cache = NULL, favicon_cache = NULL, updated_at = CURRENT_TIMESTAMP, updated_by = ? WHERE id = ?;

-- name: AdminDeleteExternalLinkByURL :exec
DELETE FROM external_links WHERE url = ?;

-- name: UpdateExternalLinkMetadata :exec
UPDATE external_links
SET card_title = ?, card_description = ?, card_image = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CreateExternalLink :execresult
INSERT INTO external_links (url, clicks)
VALUES (?, 0);
