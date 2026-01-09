-- name: CreateDraft :execlastid
INSERT INTO drafts (
    user_id, thread_id, name, content
) VALUES (?, ?, ?, ?);

-- name: UpdateDraft :exec
UPDATE drafts
SET name = ?, content = ?
WHERE id = ? AND user_id = ?;

-- name: DeleteDraft :exec
UPDATE drafts
SET deleted_at = NOW()
WHERE id = ? AND user_id = ?;

-- name: GetDraft :one
SELECT *
FROM drafts
WHERE id = ? AND user_id = ? AND deleted_at IS NULL;

-- name: ListDraftsForThread :many
SELECT *
FROM drafts
WHERE thread_id = ? AND user_id = ? AND deleted_at IS NULL;
