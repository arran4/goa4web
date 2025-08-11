-- name: AddContentPublicLabel :exec
INSERT IGNORE INTO content_public_labels (
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
INSERT IGNORE INTO content_private_labels (
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
INSERT IGNORE INTO content_label_status (
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
