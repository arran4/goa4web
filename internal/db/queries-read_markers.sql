-- name: UpsertContentReadMarker :exec
INSERT INTO content_read_markers (
    item, item_id, user_id, last_comment_id
) VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE last_comment_id = VALUES(last_comment_id);

-- name: GetContentReadMarker :one
SELECT item, item_id, user_id, last_comment_id
FROM content_read_markers
WHERE item = ? AND item_id = ? AND user_id = ?;

<<<<<<< HEAD
=======

-- name: CheckIfOthersReadCommentForPoster :one
SELECT 1 FROM content_read_markers
WHERE item = ? AND item_id = ? AND user_id != ? AND last_comment_id >= ?
LIMIT 1;
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
