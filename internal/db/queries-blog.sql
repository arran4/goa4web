-- name: UpdateBlogEntry :exec
UPDATE blogs b
SET b.language_idlanguage = ?, b.blog = ?
WHERE b.idblogs = ?
  AND EXISTS (
      SELECT 1 FROM grants g
       WHERE g.section = 'blogs'
         AND g.item = 'entry'
         AND g.action = 'edit'
         AND g.active = 1
         AND g.item_id = b.idblogs
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT rid FROM (
                SELECT ur.role_id AS rid FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
                UNION
                SELECT r2.id FROM user_roles ur2
                    JOIN grants gr ON gr.role_id = ur2.role_id AND gr.section = 'role' AND gr.active = 1
                    JOIN roles r2 ON r2.name = gr.action
                WHERE ur2.users_idusers = sqlc.arg(viewer_id)
            ) AS role_ids
        ))
  );

-- name: CreateBlogEntry :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
SELECT ?, ?, ?, CURRENT_TIMESTAMP
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND g.item = 'entry'
      AND g.action = 'post'
      AND g.active = 1
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT rid FROM (
              SELECT ur.role_id AS rid FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
              UNION
              SELECT r2.id FROM user_roles ur2
                  JOIN grants gr ON gr.role_id = ur2.role_id AND gr.section = 'role' AND gr.active = 1
                  JOIN roles r2 ON r2.name = gr.action
              WHERE ur2.users_idusers = sqlc.arg(viewer_id)
          ) AS role_ids
      ))
);

-- name: AssignThreadIdToBlogEntry :exec
UPDATE blogs
SET forumthread_id = ?
WHERE idblogs = ?;

-- name: GetBlogEntriesForUserDescending :many
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(Viewer_idusers) AS is_owner
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogEntriesByAuthorForUserDescendingLanguages :many
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(viewer_id) AS is_owner
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE (b.users_idusers = sqlc.arg(author_id) OR sqlc.arg(author_id) = 0)
AND (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogEntriesByIdsDescending :many
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE b.idblogs IN (sqlc.slice(blogIds))
ORDER BY b.written DESC
;

-- name: GetBlogEntryForUserById :one
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(Viewer_idusers) AS is_owner
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE b.idblogs = sqlc.arg(id)
LIMIT 1;

-- name: GetCountOfBlogPostsByUser :many
SELECT u.username, COUNT(b.idblogs)
FROM blogs b, users u
WHERE b.users_idusers = u.idusers
GROUP BY u.idusers;

-- name: BlogsSearchFirst :many
SELECT DISTINCT cs.blog_id
FROM blogs_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: BlogsSearchNext :many
SELECT DISTINCT cs.blog_id
FROM blogs_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.blog_id IN (sqlc.slice('ids'))
;


-- name: GetAllBlogEntriesByUser :many
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE b.users_idusers = ?
ORDER BY b.written DESC;

-- name: ListBloggersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(b.idblogs) AS count
FROM blogs b
JOIN users u ON b.users_idusers = u.idusers
WHERE (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND g.item = 'entry'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = b.idblogs
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: SearchBloggersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(b.idblogs) AS count
FROM blogs b
JOIN users u ON b.users_idusers = u.idusers
WHERE (LOWER(u.username) LIKE LOWER(sqlc.arg(query)) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(sqlc.arg(query)))
  AND (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND g.item = 'entry'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = b.idblogs
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: SetBlogLastIndex :exec
UPDATE blogs SET last_index = NOW() WHERE idblogs = ?;


-- name: GetAllBlogsForIndex :many
SELECT idblogs, blog FROM blogs WHERE deleted_at IS NULL;

