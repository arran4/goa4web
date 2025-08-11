-- name: AdminCountThreadsByBoard :one
SELECT COUNT(DISTINCT forumthread_id)
FROM imagepost
WHERE imageboard_idimageboard = ?;

-- name: AdminForumTopicThreadCounts :many
SELECT t.id, t.title, t.handler,
       COUNT(DISTINCT th.id) AS threads,
       COUNT(c.idcomments) AS comments
FROM forumtopic t
LEFT JOIN forumthread th ON th.topic_id = t.id
LEFT JOIN comments c ON c.forumthread_id = th.id
GROUP BY t.id, t.title, t.handler
ORDER BY t.title;

-- name: AdminForumCategoryThreadCounts :many
SELECT c.id, c.title,
       COUNT(DISTINCT th.id) AS threads,
       COUNT(cm.idcomments) AS comments
FROM forumcategory c
LEFT JOIN forumtopic t ON c.id = t.category_id
LEFT JOIN forumthread th ON th.topic_id = t.id
LEFT JOIN comments cm ON cm.forumthread_id = th.id
GROUP BY c.id
ORDER BY c.title;

-- name: AdminForumHandlerThreadCounts :many
SELECT t.handler,
       COUNT(DISTINCT th.id) AS threads,
       COUNT(c.idcomments) AS comments
FROM forumtopic t
LEFT JOIN forumthread th ON th.topic_id = t.id
LEFT JOIN comments c ON c.forumthread_id = th.id
GROUP BY t.handler
ORDER BY t.handler;

-- name: AdminImageboardPostCounts :many
SELECT ib.idimageboard, ib.title, COUNT(ip.idimagepost) AS count
FROM imageboard ib
LEFT JOIN imagepost ip ON ip.imageboard_idimageboard = ib.idimageboard
GROUP BY ib.idimageboard
ORDER BY ib.title;

-- name: AdminWritingCategoryCounts :many
SELECT wc.idwritingCategory, wc.title, COUNT(w.idwriting) AS count
FROM writing_category wc
LEFT JOIN writing w ON w.writing_category_id = wc.idwritingCategory
GROUP BY wc.idwritingCategory
ORDER BY wc.title;

-- name: AdminUserPostCounts :many
SELECT u.idusers, u.username,
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

-- name: AdminUserPostCountsByID :one
SELECT
  (SELECT COUNT(*) FROM blogs b WHERE b.users_idusers = u.idusers)      AS blogs,
  (SELECT COUNT(*) FROM site_news n WHERE n.users_idusers = u.idusers)  AS news,
  (SELECT COUNT(*) FROM comments c WHERE c.users_idusers = u.idusers)   AS comments,
  (SELECT COUNT(*) FROM imagepost i WHERE i.users_idusers = u.idusers)  AS images,
  (SELECT COUNT(*) FROM linker l WHERE l.users_idusers = u.idusers)     AS links,
  (SELECT COUNT(*) FROM writing w WHERE w.users_idusers = u.idusers)    AS writings
FROM users u
WHERE u.idusers = ?;

-- name: SystemGetTemplateOverride :one
SELECT body FROM template_overrides WHERE name = ?;

-- name: AdminSetTemplateOverride :exec
INSERT INTO template_overrides (name, body)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE body = VALUES(body);

-- name: AdminDeleteTemplateOverride :exec
DELETE FROM template_overrides WHERE name = ?;

-- name: SystemListUserInfo :many
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

-- name: AdminGetRecentAuditLogs :many
SELECT a.id, a.users_idusers, u.username, a.action, a.path, a.details, a.data, a.created_at
FROM audit_log a LEFT JOIN users u ON a.users_idusers = u.idusers
ORDER BY a.id DESC LIMIT ?;

-- name: AdminGetDashboardStats :one
SELECT
    (SELECT COUNT(*) FROM users) AS users,
    (SELECT COUNT(*) FROM language) AS languages,
    (SELECT COUNT(*) FROM site_news) AS news,
    (SELECT COUNT(*) FROM blogs) AS blogs,
    (SELECT COUNT(*) FROM forumtopic) AS forum_topics,
    (SELECT COUNT(*) FROM forumthread) AS forum_threads,
    (SELECT COUNT(*) FROM writing) AS writings;

-- name: AdminGetSearchStats :one
SELECT
    (SELECT COUNT(*) FROM searchwordlist) AS words,
    (SELECT COUNT(*) FROM comments_search) AS comments,
    (SELECT COUNT(*) FROM site_news_search) AS news,
    (SELECT COUNT(*) FROM blogs_search) AS blogs,
    (SELECT COUNT(*) FROM linker_search) AS linker,
    (SELECT COUNT(*) FROM writing_search) AS writings,
    (SELECT COUNT(*) FROM imagepost_search) AS images;

-- name: AdminGetForumStats :one
SELECT
    (SELECT COUNT(*) FROM forumcategory) AS categories,
    (SELECT COUNT(*) FROM forumtopic) AS topics,
    (SELECT COUNT(*) FROM forumthread) AS threads;

-- name: AdminCountForumThreads :one
SELECT COUNT(*) FROM forumthread;

-- name: AdminCountForumTopics :one
SELECT COUNT(*) FROM forumtopic;
