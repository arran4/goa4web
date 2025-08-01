-- name: AdminPromoteAnnouncement :exec
INSERT INTO site_announcements (site_news_id)
VALUES (?);

-- name: AdminDemoteAnnouncement :exec
DELETE FROM site_announcements WHERE id = ?;

-- name: GetLatestAnnouncementByNewsID :one
SELECT id, site_news_id, active, created_at
FROM site_announcements
WHERE site_news_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: SetAnnouncementActive :exec
UPDATE site_announcements SET active = ? WHERE id = ?;

-- name: GetActiveAnnouncementWithNewsForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT a.id, n.idsiteNews, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
WHERE a.active = 1
  AND (
      NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
      OR n.language_idlanguage = 0
      OR n.language_idlanguage IS NULL
      OR n.language_idlanguage IN (
          SELECT ul.language_idlanguage
          FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='news'
        AND g.item='post'
        AND g.action='view'
        AND g.active=1
        AND g.item_id = n.idsiteNews
        AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY a.created_at DESC
LIMIT 1;

-- name: AdminListAnnouncementsWithNews :many
SELECT a.id, a.site_news_id, a.active, a.created_at, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
ORDER BY a.created_at DESC;
