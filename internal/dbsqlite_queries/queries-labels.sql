-- name: AddContentPublicLabel :exec
INSERT OR IGNORE INTO content_public_labels (
    item, item_id, label
) VALUES (?, ?, ?);

-- name: RemoveContentPublicLabel :exec
DELETE FROM content_public_labels
WHERE item = ? AND item_id = ? AND label = ?;

-- name: ListContentPublicLabels :many
SELECT item, item_id, label
FROM content_public_labels
WHERE item = ? AND item_id = ?;

-- name: AddContentPrivateLabel :exec
INSERT OR IGNORE INTO content_private_labels (
    item, item_id, user_id, label, invert
) VALUES (?, ?, ?, ?, ?);

-- name: RemoveContentPrivateLabel :exec
DELETE FROM content_private_labels
WHERE item = ? AND item_id = ? AND user_id = ? AND label = ?;

-- name: ListContentPrivateLabels :many
SELECT item, item_id, user_id, label, invert
FROM content_private_labels
WHERE item = ? AND item_id = ? AND user_id = ?;

-- name: SystemClearContentPrivateLabel :exec
DELETE FROM content_private_labels
WHERE item = ? AND item_id = ? AND label = ?;

-- name: ClearUnreadContentPrivateLabelExceptUser :exec
DELETE FROM content_private_labels
WHERE item = ? AND item_id = ? AND label = 'unread' AND invert = true AND user_id <> ?;

-- name: AddContentLabelStatus :exec
INSERT OR IGNORE INTO content_label_status (
    item, item_id, label
) VALUES (?, ?, ?);

-- name: RemoveContentLabelStatus :exec
DELETE FROM content_label_status
WHERE item = ? AND item_id = ? AND label = ?;

-- name: ListContentLabelStatus :many
SELECT item, item_id, label
FROM content_label_status
WHERE item = ? AND item_id = ?;

-- name: SystemClearContentLabelStatus :exec
DELETE FROM content_label_status
WHERE item = ? AND item_id = ?;

-- name: SystemReopenThreadUnreadLabelForNewerActivity :exec
DELETE FROM content_private_labels
WHERE item = 'thread'
  AND item_id = sqlc.arg(thread_id)
  AND user_id = sqlc.arg(recipient_id)
  AND label = 'unread'
  AND invert = true
  AND EXISTS (
      SELECT 1 FROM comments c
      WHERE c.forumthread_id = content_private_labels.item_id
        AND c.deleted_at IS NULL
        AND (
            c.language_id = 0 OR c.language_id IS NULL
            OR EXISTS (
                SELECT 1 FROM user_language ul
                WHERE ul.users_idusers = content_private_labels.user_id
                  AND ul.language_id = c.language_id
            )
            OR NOT EXISTS (
                SELECT 1 FROM user_language ul
                WHERE ul.users_idusers = content_private_labels.user_id
            )
        )
        AND (
            c.idcomments > sqlc.arg(last_comment_id)
            OR (c.idcomments = sqlc.arg(last_comment_id) AND EXISTS (
                SELECT 1 FROM comment_appends ca
                WHERE ca.comment_id = c.idcomments AND ca.id > sqlc.arg(last_append_id)
            ))
        )
  );
