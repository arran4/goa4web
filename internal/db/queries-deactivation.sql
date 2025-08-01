-- Queries for user deactivation and restoration

-- name: ArchiveUserForAdmin :exec
INSERT INTO deactivated_users (idusers, email, passwd, passwd_algorithm, username, deleted_at)
SELECT u.idusers, u.email, u.passwd, u.passwd_algorithm, u.username, NOW()
FROM users u WHERE u.idusers = ?;

-- name: ScrubUserForAdmin :exec
UPDATE users SET username = ?, email = '', passwd = '', passwd_algorithm = '', deleted_at = NOW()
WHERE idusers = ?;

-- name: ArchiveCommentForAdmin :exec
INSERT INTO deactivated_comments (idcomments, forumthread_id, users_idusers, language_idlanguage, written, text, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubCommentForAdmin :exec
UPDATE comments SET text = ?, deleted_at = NOW() WHERE idcomments = ?;

-- name: PendingDeactivatedCommentsForAdmin :many
SELECT idcomments, text FROM deactivated_comments
WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreCommentForAdmin :exec
UPDATE comments SET text = ?, deleted_at = NULL WHERE idcomments = ?;

-- name: MarkCommentRestoredForAdmin :exec
UPDATE deactivated_comments SET restored_at = NOW() WHERE idcomments = ?;

-- name: ArchiveWritingForAdmin :exec
INSERT INTO deactivated_writings (idwriting, users_idusers, forumthread_id, language_idlanguage, writing_category_id, title, published, writing, abstract, private, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubWritingForAdmin :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, deleted_at = NOW() WHERE idwriting = ?;

-- name: PendingDeactivatedWritingsForAdmin :many
SELECT idwriting, title, writing, abstract, private FROM deactivated_writings
WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreWritingForAdmin :exec
UPDATE writing SET title = ?, writing = ?, abstract = ?, private = ?, deleted_at = NULL WHERE idwriting = ?;

-- name: MarkWritingRestoredForAdmin :exec
UPDATE deactivated_writings SET restored_at = NOW() WHERE idwriting = ?;

-- name: ArchiveBlogForAdmin :exec
INSERT INTO deactivated_blogs (idblogs, forumthread_id, users_idusers, language_idlanguage, blog, written, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubBlogForAdmin :exec
UPDATE blogs SET blog = ?, deleted_at = NOW() WHERE idblogs = ?;

-- name: PendingDeactivatedBlogsForAdmin :many
SELECT idblogs, blog FROM deactivated_blogs WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreBlogForAdmin :exec
UPDATE blogs SET blog = ?, deleted_at = NULL WHERE idblogs = ?;

-- name: MarkBlogRestoredForAdmin :exec
UPDATE deactivated_blogs SET restored_at = NOW() WHERE idblogs = ?;

-- name: ArchiveImagepostForAdmin :exec
INSERT INTO deactivated_imageposts (idimagepost, forumthread_id, users_idusers, imageboard_idimageboard, posted, description, thumbnail, fullimage, file_size, approved, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubImagepostForAdmin :exec
UPDATE imagepost SET description = '', thumbnail = '', fullimage = '', deleted_at = NOW() WHERE idimagepost = ?;

-- name: PendingDeactivatedImagepostsForAdmin :many
SELECT idimagepost, description, thumbnail, fullimage FROM deactivated_imageposts WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreImagepostForAdmin :exec
UPDATE imagepost SET description = ?, thumbnail = ?, fullimage = ?, deleted_at = NULL WHERE idimagepost = ?;

-- name: MarkImagepostRestoredForAdmin :exec
UPDATE deactivated_imageposts SET restored_at = NOW() WHERE idimagepost = ?;

-- name: ArchiveLinkForAdmin :exec
INSERT INTO deactivated_linker (idlinker, language_idlanguage, users_idusers, linker_category_id, forumthread_id, title, url, description, listed, deleted_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- name: ScrubLinkForAdmin :exec
UPDATE linker SET title = ?, url = '', description = '', deleted_at = NOW() WHERE idlinker = ?;

-- name: PendingDeactivatedLinksForAdmin :many
SELECT idlinker, title, url, description FROM deactivated_linker WHERE users_idusers = ? AND restored_at IS NULL;

-- name: RestoreLinkForAdmin :exec
UPDATE linker SET title = ?, url = ?, description = ?, deleted_at = NULL WHERE idlinker = ?;

-- name: MarkLinkRestoredForAdmin :exec
UPDATE deactivated_linker SET restored_at = NOW() WHERE idlinker = ?;

-- name: RestoreUserForAdmin :exec
UPDATE users u JOIN deactivated_users d ON u.idusers = d.idusers
SET u.email = d.email, u.passwd = d.passwd, u.passwd_algorithm = d.passwd_algorithm, u.username = d.username, u.deleted_at = NULL, d.restored_at = NOW()
WHERE u.idusers = ? AND d.restored_at IS NULL;
