-- name: UpsertContentReadMarker :exec
INSERT INTO content_read_markers (
    item, item_id, user_id, last_comment_id
) VALUES (?, ?, ?, ?)
ON CONFLICT(user_id, item_id, item) DO UPDATE SET last_comment_id = excluded.last_comment_id;

-- name: GetContentReadMarker :one
SELECT item, item_id, user_id, last_comment_id
FROM content_read_markers
WHERE item = ? AND item_id = ? AND user_id = ?;

-- name: SystemHasOtherUserReadItemAtOrBeyond :one
SELECT EXISTS (
    SELECT 1 FROM content_read_markers crm
    WHERE crm.item = sqlc.arg(item)
      AND crm.item_id = sqlc.arg(item_id)
      AND crm.user_id != sqlc.arg(user_id)
      AND crm.last_comment_id >= sqlc.arg(last_comment_id)
) AS has_read;
