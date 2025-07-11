-- name: GetCommentByIdForUser :one
SELECT c.*, pu.Username
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN user_roles ur ON ur.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = ? AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = t.idforumtopic
      AND (g.user_id = ur.users_idusers OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id = ur.role_id)
)
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
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN user_roles ur ON ur.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids')) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = t.idforumtopic
      AND (g.user_id = ur.users_idusers OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id = ur.role_id)
)
ORDER BY c.written DESC
;

-- name: CreateComment :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_id, text, written)
VALUES (?, ?, ?, ?, NOW() )
;

-- name: GetCommentsByThreadIdForUser :many
SELECT c.*, pu.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN user_roles ur ON ur.users_idusers = ?
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_id=? AND c.forumthread_id!=0 AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = t.idforumtopic
      AND (g.user_id = ur.users_idusers OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id = ur.role_id)
)
ORDER BY c.written;


-- name: GetAllCommentsByUser :many
SELECT c.*, th.forumtopic_idforumtopic
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id = th.idforumthread
WHERE c.users_idusers = ?
ORDER BY c.written;
