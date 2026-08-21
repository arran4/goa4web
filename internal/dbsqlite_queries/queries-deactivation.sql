-- Queries for user deactivation and restoration

-- name: AdminArchiveUser :exec
INSERT INTO deactivated_users (idusers, email, passwd, passwd_algorithm, username, deleted_at)
SELECT idusers,
       COALESCE((SELECT email FROM user_emails WHERE user_id = u.idusers ORDER BY id DESC LIMIT 1), ''),
       COALESCE((SELECT passwd FROM passwords WHERE passwords.users_idusers = u.idusers ORDER BY id DESC LIMIT 1), ''),
       COALESCE((SELECT passwd_algorithm FROM passwords WHERE passwords.users_idusers = u.idusers ORDER BY id DESC LIMIT 1), ''),
       u.username,
       CURRENT_TIMESTAMP
FROM users u WHERE u.idusers = ?;

-- name: AdminScrubUser :exec
UPDATE users SET username = ?, deleted_at = CURRENT_TIMESTAMP WHERE idusers = ?;

-- name: AdminScrubUserEmails :exec
DELETE FROM user_emails WHERE user_id = ?;

-- name: AdminScrubUserPasswords :exec
DELETE FROM passwords WHERE users_idusers = ?;

-- name: AdminArchiveComment :exec
INSERT INTO deactivated_comments (idcomments, forumthread_id, users_idusers, language_id, written, text, timezone, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: AdminScrubComment :exec
UPDATE comments SET text = ?, deleted_at = CURRENT_TIMESTAMP WHERE idcomments = ?;

-- name: AdminListPendingDeactivatedComments :many
SELECT idcomments, text FROM deactivated_comments
WHERE users_idusers = ? AND restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreComment :exec
UPDATE comments SET text = ?, deleted_at = NULL WHERE idcomments = ?;

-- name: AdminMarkCommentRestored :exec
UPDATE deactivated_comments SET restored_at = CURRENT_TIMESTAMP WHERE idcomments = ?;

-- name: AdminIsCommentDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_comments
    WHERE idcomments = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedComments :many
SELECT idcomments, text FROM deactivated_comments
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminArchiveWriting :exec
INSERT INTO deactivated_writings (idwriting, users_idusers, forumthread_id, language_id, writing_category_id, title, published, timezone, writing, abstract, private, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: AdminScrubWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, deleted_at = CURRENT_TIMESTAMP WHERE idwriting = ?;

-- name: AdminListPendingDeactivatedWritings :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE users_idusers = ? AND restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, private = ?, deleted_at = NULL WHERE idwriting = ?;

-- name: AdminMarkWritingRestored :exec
UPDATE deactivated_writings SET restored_at = CURRENT_TIMESTAMP WHERE idwriting = ?;

-- name: AdminIsWritingDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_writings
    WHERE idwriting = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedWritings :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminArchiveBlog :exec
INSERT INTO deactivated_blogs (idblogs, forumthread_id, users_idusers, language_id, blog, written, timezone, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: AdminScrubBlog :exec
UPDATE blogs SET blog = ?, deleted_at = CURRENT_TIMESTAMP WHERE idblogs = ?;

-- name: AdminListPendingDeactivatedBlogs :many
SELECT idblogs, blog FROM deactivated_blogs WHERE users_idusers = ? AND restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreBlog :exec
UPDATE blogs SET blog = ?, deleted_at = NULL WHERE idblogs = ?;

-- name: AdminMarkBlogRestored :exec
UPDATE deactivated_blogs SET restored_at = CURRENT_TIMESTAMP WHERE idblogs = ?;

-- name: AdminIsBlogDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_blogs
    WHERE idblogs = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedBlogs :many
SELECT idblogs, blog FROM deactivated_blogs
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminArchiveImagepost :exec
INSERT INTO deactivated_imageposts (idimagepost, forumthread_id, users_idusers, imageboard_idimageboard, posted, timezone, description, thumbnail, fullimage, file_size, approved, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: AdminScrubImagepost :exec
UPDATE imagepost SET description = '', thumbnail = '', fullimage = '', deleted_at = CURRENT_TIMESTAMP WHERE idimagepost = ?;

-- name: AdminListPendingDeactivatedImageposts :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts WHERE users_idusers = ? AND restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreImagepost :exec
UPDATE imagepost SET description = ?, thumbnail = ?, fullimage = ?, deleted_at = NULL WHERE idimagepost = ?;

-- name: AdminMarkImagepostRestored :exec
UPDATE deactivated_imageposts SET restored_at = CURRENT_TIMESTAMP WHERE idimagepost = ?;

-- name: AdminIsImagepostDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_imageposts
    WHERE idimagepost = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedImageposts :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminArchiveLink :exec
INSERT INTO deactivated_linker (id, language_id, author_id, category_id, thread_id, title, url, description, listed, timezone, deleted_at)
VALUES (sqlc.arg(id), sqlc.arg(language_id), sqlc.arg(author_id), sqlc.arg(category_id), sqlc.arg(thread_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description), sqlc.arg(listed), sqlc.arg(timezone), CURRENT_TIMESTAMP);

-- name: AdminScrubLink :exec
UPDATE linker SET title = ?, url = '', description = '', deleted_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: AdminListPendingDeactivatedLinks :many
SELECT id, title, url, description FROM deactivated_linker WHERE author_id = ? AND restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreLink :exec
UPDATE linker SET title = ?, url = ?, description = ?, deleted_at = NULL WHERE id = ?;

-- name: AdminMarkLinkRestored :exec
UPDATE deactivated_linker SET restored_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: AdminIsLinkDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_linker
    WHERE id = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedLinks :many
SELECT id, title, url, description FROM deactivated_linker
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminRestoreUser :exec
UPDATE users
SET username = (SELECT username FROM deactivated_users WHERE deactivated_users.idusers = users.idusers), deleted_at = NULL
WHERE users.idusers = ? AND EXISTS (SELECT 1 FROM deactivated_users d WHERE d.idusers = users.idusers AND d.restored_at IS NULL);

-- name: AdminRestoreUserEmail :exec
INSERT INTO user_emails (user_id, email, verified_at)
SELECT idusers, email, CURRENT_TIMESTAMP FROM deactivated_users
WHERE idusers = ? AND restored_at IS NULL
ON CONFLICT(user_id) DO UPDATE SET email = excluded.email;

-- name: AdminRestoreUserPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)
SELECT idusers, passwd, passwd_algorithm FROM deactivated_users
WHERE idusers = ? AND restored_at IS NULL
ON CONFLICT(users_idusers) DO UPDATE SET passwd = excluded.passwd, passwd_algorithm = excluded.passwd_algorithm;

-- name: AdminMarkUserRestored :exec
UPDATE deactivated_users SET restored_at = CURRENT_TIMESTAMP WHERE idusers = ?;

-- name: AdminIsUserDeactivated :one
SELECT EXISTS(
    SELECT 1 FROM deactivated_users
    WHERE idusers = ? AND restored_at IS NULL
) AS is_deactivated;

-- name: AdminListDeactivatedUsers :many
SELECT idusers, email, username FROM deactivated_users
WHERE restored_at IS NULL
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminGetDeactivatedCommentById :one
SELECT * FROM deactivated_comments
WHERE idcomments = ? AND restored_at IS NULL;
