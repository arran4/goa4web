-- name: GetCommentByIdForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT c.*, pu.Username,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = sqlc.arg(id) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
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
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner,
       th.idforumthread, t.idforumtopic, t.title AS forumtopic_title, fc.idforumcategory, fc.title AS forumcategory_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids')) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY c.written DESC
;

-- name: CreateComment :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_id, text, written)
VALUES (?, ?, ?, ?, NOW() )
;

-- name: GetCommentsByThreadIdForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_id=sqlc.arg(thread_id) AND c.forumthread_id!=0 AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY c.written;


-- name: AdminGetAllCommentsByUser :many
SELECT c.*, th.forumtopic_idforumtopic
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id = th.idforumthread
WHERE c.users_idusers = ?
ORDER BY c.written;

-- name: SetCommentLastIndex :exec
UPDATE comments SET last_index = NOW() WHERE idcomments = ?;


-- name: GetAllCommentsForIndex :many
SELECT idcomments, text FROM comments WHERE deleted_at IS NULL;

-- name: AdminListAllCommentsWithThreadInfo :many
SELECT c.idcomments, c.written, c.text, c.deleted_at,
       th.idforumthread, t.idforumtopic, t.title AS forumtopic_title,
       u.idusers, u.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id = th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN users u ON u.idusers = c.users_idusers
ORDER BY c.written DESC
LIMIT ? OFFSET ?;
