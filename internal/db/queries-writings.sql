-- name: GetPublicWritings :many
SELECT w.*
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: SystemListPublicWritingsByAuthor :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = sqlc.arg(author_id)
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: SystemGetWritingByID :one
SELECT forumthread_id
FROM writing
WHERE idwriting = ?;

-- name: ListPublicWritingsByUserForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = sqlc.arg(author_id)
  AND (
    w.language_idlanguage = 0
    OR w.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = w.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: SystemListPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers = u.idusers
WHERE w.private = 0 AND w.writing_category_id = sqlc.arg(category_id)
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: ListPublicWritingsInCategoryForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id = sqlc.arg(writing_category_id)
  AND (
    w.language_idlanguage = 0
    OR w.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = w.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: UpdateWritingForWriter :exec
UPDATE writing w
SET title = sqlc.arg(title),
    abstract = sqlc.arg(abstract),
    writing = sqlc.arg(content),
    private = sqlc.arg(private),
    language_idlanguage = sqlc.arg(language_id)
WHERE w.idwriting = sqlc.arg(writing_id)
  AND w.users_idusers = sqlc.arg(writer_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
  );

-- name: InsertWriting :execlastid
INSERT INTO writing (writing_category_id, title, abstract, writing, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?);

-- name: GetWritingForListerByID :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting = sqlc.arg(idwriting)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
;

-- name: ListWritingsByIDsForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting IN (sqlc.slice(writing_ids))
  AND (
    w.language_idlanguage = 0
    OR w.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = w.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: AdminInsertWritingCategory :exec
INSERT INTO writing_category (writing_category_id, title, description)
VALUES (?, ?, ?);

-- name: AdminUpdateWritingCategory :exec
UPDATE writing_category
SET title = ?, description = ?, writing_category_id = ?
WHERE idwritingCategory = ?;

-- name: SystemListWritingCategories :many
SELECT wc.*
FROM writing_category wc
ORDER BY wc.idwritingcategory
LIMIT ? OFFSET ?;

-- name: ListWritingCategoriesForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT wc.*
FROM writing_category wc
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='category' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = wc.idwritingcategory OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
);


-- name: SystemAssignWritingThreadID :exec
UPDATE writing SET forumthread_id = ? WHERE idwriting = ?;


-- name: GetAllWritingsByAuthorForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = sqlc.arg(author_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND (g.item='article' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC;

-- name: AdminGetAllWritingsByAuthor :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = sqlc.arg(author_id)
ORDER BY w.published DESC;
-- name: ListWritersForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (
    w.language_idlanguage = 0
    OR w.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = w.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND (g.item = 'article' OR g.item IS NULL)
      AND g.action = 'see'
      AND g.active = 1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: ListWritersSearchForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (LOWER(u.username) LIKE LOWER(sqlc.arg(query)) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(sqlc.arg(query)))
  AND (
    w.language_idlanguage = 0
    OR w.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(lister_id)
          AND ul.language_idlanguage = w.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND (g.item = 'article' OR g.item IS NULL)
      AND g.action = 'see'
      AND g.active = 1
      AND (g.item_id = w.idwriting OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: SystemSetWritingLastIndex :exec
UPDATE writing SET last_index = NOW() WHERE idwriting = ?;


-- name: GetAllWritingsForIndex :many
SELECT idwriting, title, abstract, writing FROM writing WHERE deleted_at IS NULL;

-- name: GetWritingCategoryById :one
SELECT * FROM writing_category WHERE idwritingCategory = ?;

-- name: AdminGetWritingsByCategoryId :many
SELECT w.*, u.username
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.writing_category_id = ?
ORDER BY w.published DESC;
