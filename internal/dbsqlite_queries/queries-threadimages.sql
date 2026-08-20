-- name: ListThreadImagePaths :many
SELECT path
FROM thread_images
WHERE forumthread_id = ?
  AND path IN (sqlc.slice(paths));

-- name: CreateThreadImage :exec
INSERT INTO thread_images (forumthread_id, path, created_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (forumthread_id, path) DO NOTHING;
