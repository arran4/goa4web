-- name: ListThreadImagePaths :many
SELECT path
FROM thread_images
WHERE forumthread_id = sqlc.arg(thread_id)
  AND path IN (sqlc.slice(paths));

-- name: CreateThreadImage :exec
INSERT INTO thread_images (forumthread_id, path, created_at)
SELECT sqlc.arg(thread_id), sqlc.arg(path), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM thread_images ti
    WHERE ti.forumthread_id = sqlc.arg(thread_id)
      AND ti.path = sqlc.arg(path)
);
