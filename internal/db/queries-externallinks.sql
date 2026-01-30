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

-- name: UpdateExternalLinkImageCache :exec
UPDATE external_links
SET card_image_cache = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CreateExternalLink :execresult
INSERT INTO external_links (url, clicks)
VALUES (?, 0);

-- name: EnsureExternalLink :execresult
INSERT INTO external_links (url, clicks)
VALUES (?, 0)
ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);

-- name: AdminGetExternalLinkByCacheID :one
SELECT * FROM external_links WHERE card_image_cache = ? OR favicon_cache = ? LIMIT 1;
