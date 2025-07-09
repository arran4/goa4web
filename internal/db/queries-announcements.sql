-- name: CreateAnnouncement :exec
INSERT INTO site_announcements (site_news_id)
VALUES (?);

-- name: DeleteAnnouncement :exec
DELETE FROM site_announcements WHERE id = ?;

-- name: GetLatestAnnouncementByNewsID :one
SELECT id, site_news_id, active, created_at
FROM site_announcements
WHERE site_news_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: SetAnnouncementActive :exec
UPDATE site_announcements SET active = ? WHERE id = ?;

-- name: GetActiveAnnouncementWithNews :one
SELECT a.id, n.idsiteNews, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
WHERE a.active = 1
ORDER BY a.created_at DESC
LIMIT 1;

-- name: ListAnnouncementsWithNews :many
SELECT a.id, a.site_news_id, a.active, a.created_at, n.news
FROM site_announcements a
JOIN site_news n ON n.idsiteNews = a.site_news_id
ORDER BY a.created_at DESC;
