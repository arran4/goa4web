-- name: RegisterExternalLinkClickSystem :exec
INSERT INTO external_links (url, clicks)
VALUES (?, 1)
ON DUPLICATE KEY UPDATE clicks = clicks + 1;

-- name: GetExternalLinkSystem :one
SELECT * FROM external_links WHERE url = ? LIMIT 1;

-- name: ListExternalLinksForAdmin :many
SELECT * FROM external_links
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetExternalLinkByIDForAdmin :one
SELECT * FROM external_links WHERE id = ? LIMIT 1;

-- name: UpdateExternalLinkForAdmin :exec
UPDATE external_links
SET url = ?, card_title = ?, card_description = ?, card_image = ?, card_image_cache = ?, favicon_cache = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ?
WHERE id = ?;

-- name: DeleteExternalLinkForAdmin :exec
DELETE FROM external_links WHERE id = ?;

-- name: ClearExternalLinkCacheForAdmin :exec
UPDATE external_links SET card_image_cache = NULL, favicon_cache = NULL, updated_at = CURRENT_TIMESTAMP, updated_by = ? WHERE id = ?;
