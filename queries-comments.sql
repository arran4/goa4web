-- name: GetCommentByIdForUser :one
SELECT c.*, pu.Username
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
LIMIT 1;

-- name: UpdateComment :exec
UPDATE comments
SET language_idlanguage = ?, text = ?
WHERE idcomments = ?;

-- name: GetCommentById :one
SELECT c.*
FROM comments c
WHERE c.Idcomments=?;

-- name: GetCommentsByIds :many
SELECT c.*
FROM comments c
WHERE c.Idcomments IN (sqlc.slice('ids'))
;

-- name: GetCommentsByIdsForUserWithThreadInfo :many
SELECT c.*, pu.username AS posterusername, th.idforumthread, t.idforumtopic, t.title AS forumtopic_title, fc.idforumcategory, fc.title AS forumcategory_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids'))
ORDER BY c.written DESC
;

-- name: CreateComment :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_idforumthread, text, written)
VALUES (?, ?, ?, ?, NOW() )
;

-- name: GetCommentsByThreadIdForUser :many
SELECT c.*, pu.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_idforumthread=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_idforumthread=? AND c.forumthread_idforumthread!=0 AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY c.written;

