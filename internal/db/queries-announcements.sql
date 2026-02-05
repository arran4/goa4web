-- name: AdminPromoteAnnouncement :exec
-- admin task
INSERT INTO site_announcements (site_news_id)
VALUES (?);

-- name: AdminDemoteAnnouncement :exec
-- admin task
DELETE FROM site_announcements WHERE id = ?;

-- name: GetLatestAnnouncementByNewsID :one
SELECT id, site_news_id, active, created_at
FROM site_announcements
WHERE site_news_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: AdminSetAnnouncementActive :exec
UPDATE site_announcements SET active = ? WHERE id = ?;

-- name: GetActiveAnnouncementWithNewsForLister :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT a.id, n.idsiteNews, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
WHERE a.active = 1
  AND (
      NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
      OR n.language_id = 0
      OR n.language_id IS NULL
      OR n.language_id IN (
          SELECT ul.language_id
          FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(lister_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='news'
        AND (g.item='post' OR g.item IS NULL)
        AND g.action='view'
        AND g.active=1
        AND (g.item_id = n.idsiteNews OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY a.created_at DESC
LIMIT 1;

-- name: AdminListAnnouncementsWithNews :many
-- admin task
SELECT a.id, a.site_news_id, a.active, a.created_at, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
ORDER BY a.created_at DESC;
