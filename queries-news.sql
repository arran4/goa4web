-- name: WriteSiteNewsRSS :many
SELECT s.idsiteNews, s.occured, s.news
FROM siteNews s
ORDER BY s.occured DESC LIMIT 15;

-- name: WriteNewsPost :exec
INSERT INTO siteNews (news, users_idusers, occured, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: EditNewsPost :exec
UPDATE siteNews SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: GetNewsThreadId :one
SELECT s.forumthread_idforumthread, u.idusers
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: AssignNewsThisThreadId :exec
UPDATE siteNews SET forumthread_idforumthread = ? WHERE idsiteNews = ?;

-- name: GetNewsPost :one
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.idsiteNews = ?
;

-- name: GetNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
;

-- name: GetLatestNewsPosts :many
SELECT u.username AS writerName, u.idusers as writerId, s.*, th.comments as Comments
FROM siteNews s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_idforumthread = th.idforumthread
ORDER BY s.occured DESC
LIMIT 15
;

