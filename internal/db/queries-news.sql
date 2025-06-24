-- name: CreateNewsPost :exec
INSERT INTO siteNews (news, users_idusers, occured, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: UpdateNewsPost :exec
UPDATE siteNews SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: GetForumThreadIdByNewsPostId :one
SELECT s.forumthread_idforumthread, u.idusers
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: AssignNewsThisThreadId :exec
UPDATE siteNews SET forumthread_idforumthread = ? WHERE idsiteNews = ?;

-- name: GetNewsPostByIdWithWriterIdAndThreadCommentCount :one
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.idsiteNews = ?
;

-- name: GetNewsPostsByIdsWithWriterIdAndThreadCommentCount :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
;

-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
ORDER BY s.occured DESC
LIMIT ? OFFSET ?
;

