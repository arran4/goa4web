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
       COUNT(DISTINCT b.idblogs) AS blogs,
       COUNT(DISTINCT n.idsiteNews) AS news,
       COUNT(DISTINCT c.idcomments) AS comments,
       COUNT(DISTINCT i.idimagepost) AS images,
       COUNT(DISTINCT l.idlinker) AS links,
       COUNT(DISTINCT w.idwriting) AS writings
FROM users u
LEFT JOIN blogs b ON b.users_idusers = u.idusers
LEFT JOIN site_news n ON n.users_idusers = u.idusers
LEFT JOIN comments c ON c.users_idusers = u.idusers
LEFT JOIN imagepost i ON i.users_idusers = u.idusers
LEFT JOIN linker l ON l.users_idusers = u.idusers
LEFT JOIN writing w ON w.users_idusers = u.idusers
GROUP BY u.idusers
ORDER BY u.username;

-- name: GetTemplateOverride :one
SELECT body FROM template_overrides WHERE name = ?;

-- name: SetTemplateOverride :exec
INSERT INTO template_overrides (name, body)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE body = VALUES(body);

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
