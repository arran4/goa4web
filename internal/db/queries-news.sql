-- name: CreateNewsPost :exec
INSERT INTO site_news (news, users_idusers, occurred, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: UpdateNewsPost :exec
UPDATE site_news SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: DeactivateNewsPost :exec
UPDATE site_news SET deleted_at = NOW() WHERE idsiteNews = ?;

-- name: GetForumThreadIdByNewsPostId :one
SELECT s.forumthread_idforumthread, u.idusers
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: AssignNewsThisThreadId :exec
UPDATE site_news SET forumthread_idforumthread = ? WHERE idsiteNews = ?;

-- name: GetNewsPostByIdWithWriterIdAndThreadCommentCount :one
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.idsiteNews = ?
;

-- name: GetNewsPostsByIdsWithWriterIdAndThreadCommentCount :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
;

-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
ORDER BY s.occurred DESC
LIMIT ? OFFSET ?
;

