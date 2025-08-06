-- name: UpdateBlogEntryForWriter :exec
UPDATE blogs b
SET language_idlanguage = sqlc.arg(language_id), blog = sqlc.arg(blog)
WHERE b.idblogs = sqlc.arg(entry_id)
  AND b.users_idusers = sqlc.arg(writer_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = 'blogs'
        AND (g.item = 'entry' OR g.item IS NULL)
        AND g.action = 'post'
        AND g.active = 1
        AND (g.item_id = sqlc.arg(grant_entry_id) OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
        ))
  );

-- name: CreateBlogEntryForWriter :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
SELECT sqlc.arg(users_idusers), sqlc.arg(language_idlanguage), sqlc.arg(blog), CURRENT_TIMESTAMP
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND (g.item = 'entry' OR g.item IS NULL)
      AND g.action = 'post'
      AND g.active = 1
      AND (g.item_id = 0 OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
      ))
);

-- name: SystemAssignBlogEntryThreadID :exec
UPDATE blogs
SET forumthread_id = ?
WHERE idblogs = ?;

-- name: SystemGetBlogEntryByID :one
SELECT idblogs, forumthread_id
FROM blogs
WHERE idblogs = ?;

-- name: ListBlogEntriesForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(lister_id) AS is_owner
FROM blogs b
JOIN grants g ON (g.item_id = b.idblogs OR g.item_id IS NULL)
    AND g.section = 'blogs'
    AND (g.item = 'entry' OR g.item IS NULL)
    AND g.action = 'see'
    AND g.active = 1
    AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
    AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: ListBlogEntriesByAuthorForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(lister_id) AS is_owner
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE (b.users_idusers = sqlc.arg(author_id) OR sqlc.arg(author_id) = 0)
AND (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND (g.item = 'entry' OR g.item IS NULL)
      AND g.action = 'see'
      AND g.active = 1
      AND (g.item_id = b.idblogs OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: ListBlogEntriesByIDsForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written
FROM blogs b
WHERE b.idblogs IN (sqlc.slice(blogIds))
  AND (
      b.language_idlanguage = 0
      OR b.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_idlanguage = b.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = 'blogs'
        AND (g.item = 'entry' OR g.item IS NULL)
        AND g.action = 'see'
        AND g.active = 1
        AND (g.item_id = b.idblogs OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogEntryForListerByID :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0),
       b.users_idusers = sqlc.arg(lister_id) AS is_owner
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE b.idblogs = sqlc.arg(id)
  AND (
      b.language_idlanguage = 0
      OR b.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_idlanguage = b.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = 'blogs'
        AND (g.item = 'entry' OR g.item IS NULL)
        AND g.action = 'see'
        AND g.active = 1
        AND (g.item_id = b.idblogs OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT 1;

-- name: ListBlogIDsBySearchWordFirstForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT DISTINCT cs.blog_id
FROM blogs_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN blogs b ON b.idblogs = cs.blog_id
WHERE swl.word = ?
  AND (
      b.language_idlanguage = 0
      OR b.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_idlanguage = b.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = 'blogs'
        AND (g.item = 'entry' OR g.item IS NULL)
        AND g.action = 'see'
        AND g.active = 1
        AND (g.item_id = b.idblogs OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: ListBlogIDsBySearchWordNextForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT DISTINCT cs.blog_id
FROM blogs_search cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist = cs.searchwordlist_idsearchwordlist
JOIN blogs b ON b.idblogs = cs.blog_id
WHERE swl.word = ?
  AND cs.blog_id IN (sqlc.slice('ids'))
  AND (
      b.language_idlanguage = 0
      OR b.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
            AND ul.language_idlanguage = b.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section = 'blogs'
        AND (g.item = 'entry' OR g.item IS NULL)
        AND g.action = 'see'
        AND g.active = 1
        AND (g.item_id = b.idblogs OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );



-- name: ListBloggersForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT u.username, COUNT(b.idblogs) AS count
FROM blogs b
JOIN users u ON b.users_idusers = u.idusers
WHERE (
    b.language_idlanguage = 0
    OR b.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND (g.item = 'entry' OR g.item IS NULL)
      AND g.action = 'see'
      AND g.active = 1
      AND (g.item_id = b.idblogs OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: ListBloggersSearchForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
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
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = b.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'blogs'
      AND (g.item = 'entry' OR g.item IS NULL)
      AND g.action = 'see'
      AND g.active = 1
      AND (g.item_id = b.idblogs OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;
-- name: AdminGetAllBlogEntriesByUser :many
SELECT b.idblogs, b.forumthread_id, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers = u.idusers
LEFT JOIN forumthread th ON b.forumthread_id = th.idforumthread
WHERE b.users_idusers = sqlc.arg(author_id)
ORDER BY b.written DESC;

-- name: SystemSetBlogLastIndex :exec
UPDATE blogs SET last_index = NOW() WHERE idblogs = sqlc.arg(id);

-- name: SystemGetAllBlogsForIndex :many
SELECT idblogs, blog FROM blogs WHERE deleted_at IS NULL;

