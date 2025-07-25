-- name: CreateNewsPost :execlastid
INSERT INTO site_news (news, users_idusers, occurred, language_idlanguage)
VALUES (?, ?, NOW(), ?);

-- name: UpdateNewsPost :exec
UPDATE site_news SET news = ?, language_idlanguage = ? WHERE idsiteNews = ?;

-- name: DeactivateNewsPost :exec
UPDATE site_news SET deleted_at = NOW() WHERE idsiteNews = ?;

-- name: GetForumThreadIdByNewsPostId :one
SELECT s.forumthread_id, u.idusers
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: AssignNewsThisThreadId :exec
UPDATE site_news SET forumthread_id = ? WHERE idsiteNews = ?;

-- name: GetNewsPostByIdWithWriterIdAndThreadCommentCount :one

WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.idsiteNews = sqlc.arg(id) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND g.item='post'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = s.idsiteNews
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
LIMIT 1;

-- name: GetNewsPostsByIdsWithWriterIdAndThreadCommentCount :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds)) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND g.item='post'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = s.idsiteNews
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY s.occurred DESC;

-- name: GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds)) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND g.item='post'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = s.idsiteNews
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY s.occurred DESC;

-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_idlanguage, s.users_idusers, s.news, s.occurred, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND g.item='post'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = s.idsiteNews
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY s.occurred DESC
LIMIT ? OFFSET ?;


-- name: SetSiteNewsLastIndex :exec
UPDATE site_news SET last_index = NOW() WHERE idsiteNews = ?;


-- name: GetAllSiteNewsForIndex :many
SELECT idsiteNews, news FROM site_news WHERE deleted_at IS NULL;

