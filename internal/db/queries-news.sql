-- name: CreateNewsPostForWriter :execlastid
INSERT INTO site_news (news, users_idusers, occurred, timezone, language_id)
SELECT sqlc.arg(news), sqlc.arg(writer_id), sqlc.arg(occurred), sqlc.arg(timezone), sqlc.narg(language_id)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND (g.item='post' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = 0 OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
);

-- name: UpdateNewsPostForWriter :exec
UPDATE site_news s
SET news = sqlc.arg(news), language_id = sqlc.narg(language_id)
WHERE s.idsiteNews = sqlc.arg(post_id)
  AND s.users_idusers = sqlc.arg(writer_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND (g.item='post' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = sqlc.arg(grant_post_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
  );

-- name: DeactivateNewsPost :exec
UPDATE site_news SET deleted_at = NOW() WHERE idsiteNews = ?;

-- name: GetForumThreadIdByNewsPostId :one
SELECT s.forumthread_id, u.idusers
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
WHERE s.idsiteNews = ?;

-- name: SystemAssignNewsThreadID :exec
UPDATE site_news SET forumthread_id = ? WHERE idsiteNews = ?;

-- name: SystemGetNewsPostByID :one
SELECT forumthread_id
FROM site_news
WHERE idsiteNews = ?;

-- name: GetNewsPostByIdWithWriterIdAndThreadCommentCount :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_id, s.users_idusers, s.news, s.occurred, s.timezone, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.idsiteNews = sqlc.arg(id) AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND (g.item='post' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = s.idsiteNews OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
LIMIT 1;


-- name: GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_id, s.users_idusers, s.news, s.occurred, s.timezone, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE s.Idsitenews IN (sqlc.slice(newsIds))
  AND (
      NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
      OR s.language_id = 0
      OR s.language_id IS NULL
      OR s.language_id IN (
          SELECT ul.language_id
          FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND (g.item='post' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = s.idsiteNews OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY s.occurred DESC;

-- name: GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_id, s.users_idusers, s.news, s.occurred, s.timezone, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
WHERE (
    NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
    OR s.language_id = 0
    OR s.language_id IS NULL
    OR s.language_id IN (
        SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='news'
      AND (g.item='post' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = s.idsiteNews OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY s.occurred DESC
LIMIT ? OFFSET ?;


-- name: AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescending :many
SELECT u.username AS writerName, u.idusers as writerId, s.idsiteNews, s.forumthread_id, s.language_id, s.users_idusers,
s.news, s.occurred, s.timezone, th.comments as Comments
FROM site_news s
LEFT JOIN users u ON s.users_idusers = u.idusers
LEFT JOIN forumthread th ON s.forumthread_id = th.idforumthread
ORDER BY s.occurred DESC
LIMIT ? OFFSET ?;


-- name: SystemSetSiteNewsLastIndex :exec
UPDATE site_news SET last_index = NOW() WHERE idsiteNews = ?;


-- name: GetAllSiteNewsForIndex :many
SELECT idsiteNews, news FROM site_news WHERE deleted_at IS NULL;

-- name: AdminReplaceSiteNewsURL :exec
UPDATE site_news SET news = REPLACE(news, sqlc.arg(old_url), sqlc.arg(new_url)) WHERE idsiteNews = sqlc.arg(id);

