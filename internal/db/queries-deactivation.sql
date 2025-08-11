-- Queries for user deactivation and restoration

-- name: AdminArchiveUser :exec
INSERT INTO deactivated_users (idusers, email, passwd, passwd_algorithm, username, deleted_at)
SELECT u.idusers, u.email, u.passwd, u.passwd_algorithm, u.username, NOW()
FROM users u WHERE u.idusers = ?;

-- name: AdminScrubUser :exec
UPDATE users SET username = ?, email = '', passwd = '', passwd_algorithm = '', deleted_at = NOW()
WHERE idusers = ?;

-- name: AdminArchiveComment :exec
INSERT INTO deactivated_comments (idcomments, forumthread_id, users_idusers, language_id, written, text, timezone, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW());

-- name: AdminScrubComment :exec
UPDATE comments SET text = ?, deleted_at = NOW() WHERE idcomments = ?;

-- name: AdminListPendingDeactivatedComments :many
SELECT idcomments, text FROM deactivated_comments
WHERE users_idusers = ? AND restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreComment :exec
UPDATE comments SET text = ?, deleted_at = NULL WHERE idcomments = ?;

-- name: AdminMarkCommentRestored :exec
UPDATE deactivated_comments SET restored_at = NOW() WHERE idcomments = ?;

-- name: AdminIsCommentDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_comments
    WHERE idcomments = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedComments :many
SELECT idcomments, text FROM deactivated_comments
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminArchiveWriting :exec
INSERT INTO deactivated_writings (idwriting, users_idusers, forumthread_id, language_id, writing_category_id, title, published, timezone, writing, abstract, private, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: AdminScrubWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, deleted_at = NOW() WHERE idwriting = ?;

-- name: AdminListPendingDeactivatedWritings :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE users_idusers = ? AND restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, private = ?, deleted_at = NULL WHERE idwriting = ?;

-- name: AdminMarkWritingRestored :exec
UPDATE deactivated_writings SET restored_at = NOW() WHERE idwriting = ?;

-- name: AdminIsWritingDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_writings
    WHERE idwriting = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedWritings :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminArchiveBlog :exec
INSERT INTO deactivated_blogs (idblogs, forumthread_id, users_idusers, language_id, blog, written, timezone, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW());

-- name: AdminScrubBlog :exec
UPDATE blogs SET blog = ?, deleted_at = NOW() WHERE idblogs = ?;

-- name: AdminListPendingDeactivatedBlogs :many
SELECT idblogs, blog FROM deactivated_blogs WHERE users_idusers = ? AND restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreBlog :exec
UPDATE blogs SET blog = ?, deleted_at = NULL WHERE idblogs = ?;

-- name: AdminMarkBlogRestored :exec
UPDATE deactivated_blogs SET restored_at = NOW() WHERE idblogs = ?;

-- name: AdminIsBlogDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_blogs
    WHERE idblogs = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedBlogs :many
SELECT idblogs, blog FROM deactivated_blogs
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminArchiveImagepost :exec
INSERT INTO deactivated_imageposts (idimagepost, forumthread_id, users_idusers, imageboard_idimageboard, posted, timezone, description, thumbnail, fullimage, file_size, approved, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: AdminScrubImagepost :exec
UPDATE imagepost SET description = '', thumbnail = '', fullimage = '', deleted_at = NOW() WHERE idimagepost = ?;

-- name: AdminListPendingDeactivatedImageposts :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts WHERE users_idusers = ? AND restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreImagepost :exec
UPDATE imagepost SET description = ?, thumbnail = ?, fullimage = ?, deleted_at = NULL WHERE idimagepost = ?;

-- name: AdminMarkImagepostRestored :exec
UPDATE deactivated_imageposts SET restored_at = NOW() WHERE idimagepost = ?;

-- name: AdminIsImagepostDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_imageposts
    WHERE idimagepost = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedImageposts :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminArchiveLink :exec
INSERT INTO deactivated_linker (id, language_id, author_id, category_id, thread_id, title, url, description, listed, timezone, deleted_at)
VALUES (sqlc.arg(id), sqlc.arg(language_id), sqlc.arg(author_id), sqlc.arg(category_id), sqlc.arg(thread_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description), sqlc.arg(listed), sqlc.arg(timezone), NOW());

-- name: AdminScrubLink :exec
UPDATE linker SET title = ?, url = '', description = '', deleted_at = NOW() WHERE id = ?;

-- name: AdminListPendingDeactivatedLinks :many
SELECT id, title, url, description FROM deactivated_linker WHERE author_id = ? AND restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreLink :exec
UPDATE linker SET title = ?, url = ?, description = ?, deleted_at = NULL WHERE id = ?;

-- name: AdminMarkLinkRestored :exec
UPDATE deactivated_linker SET restored_at = NOW() WHERE id = ?;

-- name: AdminIsLinkDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_linker
    WHERE id = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedLinks :many
SELECT id, title, url, description FROM deactivated_linker
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;

-- name: AdminRestoreUser :exec
UPDATE users u JOIN deactivated_users d ON u.idusers = d.idusers
SET u.email = d.email, u.passwd = d.passwd, u.passwd_algorithm = d.passwd_algorithm, u.username = d.username, u.deleted_at = NULL, d.restored_at = NOW()
WHERE u.idusers = ? AND d.restored_at IS NULL;

-- name: AdminIsUserDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_users
    WHERE idusers = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedUsers :many
SELECT idusers, email, username FROM deactivated_users
WHERE restored_at IS NULL
LIMIT ? OFFSET ?;
