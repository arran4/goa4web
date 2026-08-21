-- name: UpsertContentReadMarker :exec
INSERT INTO content_read_markers (
    item, item_id, user_id, last_comment_id
) VALUES (?, ?, ?, ?)
ON CONFLICT(user_id, content_id, content_type) DO UPDATE SET last_comment_id = excluded.last_comment_id;

-- name: GetContentReadMarker :one
SELECT item, item_id, user_id, last_comment_id
FROM content_read_markers
WHERE item = ? AND item_id = ? AND user_id = ?;
