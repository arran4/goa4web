-- Queries for user deactivation and restoration

-- name: ArchiveUser :exec
INSERT INTO deactivated_users (idusers, email, passwd, passwd_algorithm, username, deleted_at)
SELECT u.idusers, u.email, u.passwd, u.passwd_algorithm, u.username, NOW()
FROM users u WHERE u.idusers = ?;

-- name: ScrubUser :exec
UPDATE users SET username = ?, email = '', passwd = '', passwd_algorithm = '', deleted_at = NOW()
WHERE idusers = ?;

-- name: ArchiveComment :exec
INSERT INTO deactivated_comments (idcomments, forumthread_idforumthread, users_idusers, language_idlanguage, written, text, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubComment :exec
UPDATE comments SET text = ?, deleted_at = NOW() WHERE idcomments = ?;

-- name: PendingDeactivatedComments :many
SELECT idcomments, text FROM deactivated_comments
WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreComment :exec
UPDATE comments SET text = ?, deleted_at = NULL WHERE idcomments = ?;

-- name: MarkCommentRestored :exec
UPDATE deactivated_comments SET restored_at = NOW() WHERE idcomments = ?;

-- name: ArchiveWriting :exec
INSERT INTO deactivated_writings (idwriting, users_idusers, forumthread_idforumthread, language_idlanguage, writing_category_id, title, published, writing, abstract, private, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, deleted_at = NOW() WHERE idwriting = ?;

-- name: PendingDeactivatedWritings :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreWriting :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, private = ?, deleted_at = NULL WHERE idwriting = ?;

-- name: MarkWritingRestored :exec
UPDATE deactivated_writings SET restored_at = NOW() WHERE idwriting = ?;

-- name: ArchiveBlog :exec
INSERT INTO deactivated_blogs (idblogs, forumthread_idforumthread, users_idusers, language_idlanguage, blog, written, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubBlog :exec
UPDATE blogs SET blog = ?, deleted_at = NOW() WHERE idblogs = ?;

-- name: PendingDeactivatedBlogs :many
SELECT idblogs, blog FROM deactivated_blogs WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreBlog :exec
UPDATE blogs SET blog = ?, deleted_at = NULL WHERE idblogs = ?;

-- name: MarkBlogRestored :exec
UPDATE deactivated_blogs SET restored_at = NOW() WHERE idblogs = ?;

-- name: ArchiveImagepost :exec
INSERT INTO deactivated_imageposts (idimagepost, forumthread_idforumthread, users_idusers, imageboard_idimageboard, posted, description, thumbnail, fullimage, file_size, approved, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubImagepost :exec
UPDATE imagepost SET description = '', thumbnail = '', fullimage = '', deleted_at = NOW() WHERE idimagepost = ?;

-- name: PendingDeactivatedImageposts :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreImagepost :exec
UPDATE imagepost SET description = ?, thumbnail = ?, fullimage = ?, deleted_at = NULL WHERE idimagepost = ?;

-- name: MarkImagepostRestored :exec
UPDATE deactivated_imageposts SET restored_at = NOW() WHERE idimagepost = ?;

-- name: ArchiveLink :exec
INSERT INTO deactivated_linker (idlinker, language_idlanguage, users_idusers, linkerCategory_idlinkerCategory, forumthread_idforumthread, title, url, description, listed, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubLink :exec
UPDATE linker SET title = ?, url = '', description = '', deleted_at = NOW() WHERE idlinker = ?;

-- name: PendingDeactivatedLinks :many
SELECT idlinker, title, url, description FROM deactivated_linker WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreLink :exec
UPDATE linker SET title = ?, url = ?, description = ?, deleted_at = NULL WHERE idlinker = ?;

-- name: MarkLinkRestored :exec
UPDATE deactivated_linker SET restored_at = NOW() WHERE idlinker = ?;

-- name: RestoreUser :exec
UPDATE users u JOIN deactivated_users d ON u.idusers = d.idusers
SET u.email = d.email, u.passwd = d.passwd, u.passwd_algorithm = d.passwd_algorithm, u.username = d.username, u.deleted_at = NULL, d.restored_at = NOW()
WHERE u.idusers = ? AND d.restored_at IS NULL;
