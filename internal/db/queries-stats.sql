-- name: CountThreadsByBoard :one
SELECT COUNT(DISTINCT forumthread_id)
FROM imagepost
WHERE imageboard_idimageboard = ?;

-- name: ForumTopicThreadCounts :many
SELECT t.title, COUNT(th.idforumthread) AS count
FROM forumtopic t
LEFT JOIN forumthread th ON th.forumtopic_idforumtopic = t.idforumtopic
GROUP BY t.idforumtopic
ORDER BY t.title;

-- name: ForumCategoryThreadCounts :many
SELECT c.title, COUNT(th.idforumthread) AS count
FROM forumcategory c
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
LEFT JOIN forumthread th ON th.forumtopic_idforumtopic = t.idforumtopic
GROUP BY c.idforumcategory
ORDER BY c.title;

-- name: ImageboardPostCounts :many
SELECT ib.title, COUNT(ip.idimagepost) AS count
FROM imageboard ib
LEFT JOIN imagepost ip ON ip.imageboard_idimageboard = ib.idimageboard
GROUP BY ib.idimageboard
ORDER BY ib.title;

-- name: WritingCategoryCounts :many
SELECT wc.title, COUNT(w.idwriting) AS count
FROM writing_category wc
LEFT JOIN writing w ON w.writing_category_id = wc.idwritingCategory
GROUP BY wc.idwritingCategory
ORDER BY wc.title;

-- name: UserPostCounts :many
SELECT u.username,
       COALESCE(b.blogs, 0)     AS blogs,
       COALESCE(n.news, 0)      AS news,
       COALESCE(c.comments, 0)  AS comments,
       COALESCE(i.images, 0)    AS images,
       COALESCE(l.links, 0)     AS links,
       COALESCE(w.writings, 0)  AS writings
FROM users u
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS blogs FROM blogs GROUP BY users_idusers) b ON b.uid = u.idusers
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS news FROM site_news GROUP BY users_idusers) n ON n.uid = u.idusers
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS comments FROM comments GROUP BY users_idusers) c ON c.uid = u.idusers
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS images FROM imagepost GROUP BY users_idusers) i ON i.uid = u.idusers
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS links FROM linker GROUP BY users_idusers) l ON l.uid = u.idusers
LEFT JOIN (SELECT users_idusers AS uid, COUNT(*) AS writings FROM writing GROUP BY users_idusers) w ON w.uid = u.idusers
ORDER BY u.username;

-- name: GetTemplateOverride :one
SELECT body FROM template_overrides WHERE name = ?;

-- name: SetTemplateOverride :exec
INSERT INTO template_overrides (name, body)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE body = VALUES(body);

-- name: DeleteTemplateOverride :exec
DELETE FROM template_overrides WHERE name = ?;

-- name: ListUserInfo :many
SELECT u.idusers, u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
       IF(r.id IS NULL, 0, 1) AS admin,
       MIN(s.created_at) AS created_at
FROM users u
LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_admin = 1
LEFT JOIN sessions s ON s.users_idusers = u.idusers
GROUP BY u.idusers
ORDER BY u.idusers;

-- name: GetRecentAuditLogs :many
SELECT a.id, a.users_idusers, u.username, a.action, a.path, a.details, a.data, a.created_at
FROM audit_log a LEFT JOIN users u ON a.users_idusers = u.idusers
ORDER BY a.id DESC LIMIT ?;
