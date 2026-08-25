-- name: GetCommentByIdForUser :one
WITH user_lang AS (
    SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
),
role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT c.*, pu.Username,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.idcomments = sqlc.arg(id)
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR c.language_id IN (SELECT language_id FROM user_lang) OR (SELECT COUNT(*) FROM user_lang) = 0
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE ((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      t.handler IS NULL
      OR t.handler != 'private'
      OR EXISTS (
          SELECT 1 FROM grants thread_grant
          WHERE thread_grant.section = 'privateforum_thread'
            AND thread_grant.item = 'thread'
            AND thread_grant.action = 'view'
            AND thread_grant.active = 1
            AND thread_grant.item_id = th.idforumthread
            AND (thread_grant.user_id = sqlc.arg(user_id) OR thread_grant.user_id IS NULL)
            AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
      )
  )
LIMIT 1;

-- name: UpdateCommentForEditor :exec
UPDATE comments
SET language_id = sqlc.arg(language_id), text = sqlc.arg(text)
WHERE comments.idcomments = sqlc.arg(comment_id)
  AND comments.users_idusers = sqlc.arg(commenter_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE (g.section='forum' OR g.section='privateforum')
        AND (
            g.item IS NULL OR
            (g.item='thread' AND (g.item_id = comments.forumthread_id OR g.item_id IS NULL)) OR
            (g.item='comment' AND (g.item_id = comments.idcomments OR g.item_id IS NULL))
        )
        AND g.action='edit'
        AND g.active=1
        AND (g.user_id = sqlc.arg(editor_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(commenter_id)
        ))
  );

-- name: GetCommentById :one
SELECT c.*
FROM comments c
WHERE c.Idcomments=?;


-- name: GetCommentsByIdsForUserWithThreadInfo :many
WITH user_lang AS (
    SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
),
role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner,
       th.idforumthread, t.idforumtopic, t.title AS forumtopic_title,
       fp.text AS thread_title, fc.idforumcategory, fc.title AS forumcategory_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN comments fp ON th.firstpost = fp.idcomments
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.Idcomments IN (sqlc.slice('ids'))
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR c.language_id IN (SELECT language_id FROM user_lang) OR (SELECT COUNT(*) FROM user_lang) = 0
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE ((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      t.handler IS NULL
      OR t.handler != 'private'
      OR EXISTS (
          SELECT 1 FROM grants thread_grant
          WHERE thread_grant.section = 'privateforum_thread'
            AND thread_grant.item = 'thread'
            AND thread_grant.action = 'view'
            AND thread_grant.active = 1
            AND thread_grant.item_id = th.idforumthread
            AND (thread_grant.user_id = sqlc.arg(user_id) OR thread_grant.user_id IS NULL)
            AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
      )
  )
ORDER BY c.written DESC
;

-- name: CreateCommentInSectionForCommenter :execlastid
INSERT INTO comments (language_id, users_idusers, forumthread_id, text, written, timezone)
SELECT sqlc.arg(language_id), sqlc.arg(commenter_id), sqlc.arg(forumthread_id), sqlc.arg(text), sqlc.arg(written), sqlc.arg(timezone)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = sqlc.arg(section)
      AND (g.item = sqlc.arg(item_type) OR g.item IS NULL)
      AND g.action = sqlc.arg(action)
      AND g.active = 1
      AND (g.item_id = sqlc.arg(item_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(commenter_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(commenter_id)
      ))
  );

-- name: GetCommentsByThreadIdForUser :many
WITH user_lang AS (
    SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
),
role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_id=sqlc.arg(thread_id)
  AND c.forumthread_id IS NOT NULL
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR c.language_id IN (SELECT language_id FROM user_lang) OR (SELECT COUNT(*) FROM user_lang) = 0
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE ((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      t.handler IS NULL
      OR t.handler != 'private'
      OR EXISTS (
          SELECT 1 FROM grants thread_grant
          WHERE thread_grant.section = 'privateforum_thread'
            AND thread_grant.item = 'thread'
            AND thread_grant.action = 'view'
            AND thread_grant.active = 1
            AND thread_grant.item_id = th.idforumthread
            AND (thread_grant.user_id = sqlc.arg(user_id) OR thread_grant.user_id IS NULL)
            AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
      )
  )
ORDER BY c.written;

-- Viewing comments in a section-specific thread requires 'view' on the
-- section's primary item type since comments inherit their thread's grants.
-- name: GetCommentsBySectionThreadIdForUser :many
WITH user_lang AS (
    SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
),
role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT c.*, pu.username AS posterusername,
       c.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id=th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users pu ON pu.idusers = c.users_idusers
WHERE c.forumthread_id=sqlc.arg(thread_id)
  AND c.forumthread_id IS NOT NULL
  AND (
      c.language_id = 0
      OR c.language_id IS NULL
      OR c.language_id IN (SELECT language_id FROM user_lang) OR (SELECT COUNT(*) FROM user_lang) = 0
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section=sqlc.arg(section)
      AND (g.item=sqlc.arg(item_type) OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      t.handler IS NULL
      OR t.handler != 'private'
      OR EXISTS (
          SELECT 1 FROM grants thread_grant
          WHERE thread_grant.section = 'privateforum_thread'
            AND thread_grant.item = 'thread'
            AND thread_grant.action = 'view'
            AND thread_grant.active = 1
            AND thread_grant.item_id = th.idforumthread
            AND (thread_grant.user_id = sqlc.arg(user_id) OR thread_grant.user_id IS NULL)
            AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
      )
  )
ORDER BY c.written;


-- name: AdminGetAllCommentsByUser :many
SELECT c.idcomments, c.forumthread_id, c.users_idusers, c.language_id,
       c.written, c.text, c.deleted_at, c.last_index, c.timezone,
       th.forumtopic_idforumtopic, t.title AS forumtopic_title,
       fp.text AS thread_title
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id = th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN comments fp ON th.firstpost = fp.idcomments
WHERE c.users_idusers = sqlc.arg('user_id')
ORDER BY c.written;

-- name: SystemSetCommentLastIndex :exec
UPDATE comments SET last_index = CURRENT_TIMESTAMP WHERE idcomments = ?;

-- name: SystemListCommentsByThreadID :many
SELECT c.idcomments, c.text
FROM comments c
WHERE c.forumthread_id = ?
ORDER BY c.idcomments;


-- name: GetAllCommentsForIndex :many
SELECT idcomments, text FROM comments WHERE deleted_at IS NULL;

-- name: AdminListAllCommentsWithThreadInfo :many
SELECT c.idcomments, c.written, c.text, c.deleted_at, c.timezone,
       th.idforumthread, t.idforumtopic, t.title AS forumtopic_title, t.handler AS topic_handler,
       u.idusers, u.username AS posterusername
FROM comments c
LEFT JOIN forumthread th ON c.forumthread_id = th.idforumthread
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN users u ON u.idusers = c.users_idusers
ORDER BY c.written DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);


-- name: AppendCommentInSectionForCommenter :execrows
UPDATE comments
SET text = COALESCE(comments.text, '') || char(10) || char(10) || '[hr]' || char(10) || char(10) || CAST(sqlc.arg(text) AS TEXT), written = sqlc.arg(written)
WHERE comments.idcomments = sqlc.arg(comment_id)
  AND comments.users_idusers = sqlc.arg(commenter_id)
  AND comments.forumthread_id = sqlc.arg(forumthread_id)
  AND comments.written >= datetime('now', '-' || CAST(sqlc.arg(append_window_mins) AS INTEGER) || ' minutes')
  AND NOT EXISTS (
      SELECT 1 FROM comments newer
      WHERE newer.forumthread_id = comments.forumthread_id
        AND newer.idcomments > comments.idcomments
  )
  AND NOT EXISTS (
      SELECT 1 FROM content_read_markers crm
      WHERE crm.item = 'thread'
        AND crm.item_id = comments.forumthread_id
        AND crm.user_id != sqlc.arg(commenter_id)
        AND crm.last_comment_id >= comments.idcomments
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = sqlc.arg(section)
        AND (g.item = sqlc.arg(item_type) OR g.item IS NULL)
        AND g.action = 'append'
        AND g.active = 1
        AND (g.item_id = sqlc.arg(item_id) OR g.item_id IS NULL)
        AND (g.user_id = sqlc.narg(grant_user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(commenter_id)
        ))
  );
