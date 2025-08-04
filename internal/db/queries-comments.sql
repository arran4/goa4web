-- name: GetCommentByIdForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT c.*, pu.Username,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = sqlc.arg(id)
  AND (
      c.language_idlanguage = 0
      OR c.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = c.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
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

-- name: UpdateCommentForCommenter :exec
UPDATE comments c
SET language_idlanguage = sqlc.arg(language_id), text = sqlc.arg(text)
WHERE c.idcomments = sqlc.arg(comment_id)
  AND c.users_idusers = sqlc.arg(commenter_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='forum'
        AND (g.item='comment' OR g.item IS NULL)
        AND g.action='post'
        AND g.active=1
        AND (g.item_id = sqlc.arg(grant_comment_id) OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(commenter_id)
        ))
  );

-- name: GetCommentById :one
SELECT c.*
FROM comments c
WHERE c.Idcomments=?;


-- name: GetCommentsByIdsForUserWithThreadInfo :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner,
       th.idforumthread, t.idforumtopic, t.title AS forumtopic_title, fc.idforumcategory, fc.title AS forumcategory_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids'))
  AND (
      c.language_idlanguage = 0
      OR c.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = c.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
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

-- name: CreateCommentForCommenter :execlastid
INSERT INTO comments (language_idlanguage, users_idusers, forumthread_id, text, written)
SELECT sqlc.arg(language_id), sqlc.arg(commenter_id), sqlc.arg(forumthread_id), sqlc.arg(text), NOW()
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'forum'
      AND (g.item = 'comment' OR g.item IS NULL)
      AND g.action = 'post'
      AND g.active = 1
      AND (g.item_id = sqlc.arg(grant_forumthread_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(commenter_id)
      ))
);

-- name: GetCommentsByThreadIdForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_id=sqlc.arg(thread_id)
  AND c.forumthread_id!=0
  AND (
      c.language_idlanguage = 0
      OR c.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = c.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
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
WHERE c.users_idusers = sqlc.arg('user_id')
ORDER BY c.written;

-- name: SystemSetCommentLastIndex :exec
UPDATE comments SET last_index = NOW() WHERE idcomments = ?;

-- name: SystemListCommentsByThreadID :many
SELECT c.idcomments, c.text
FROM comments c
WHERE c.forumthread_id = ?
ORDER BY c.idcomments;


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
