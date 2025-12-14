-- name: AdminListAllPrivateTopics :many
SELECT idforumtopic, title, created_at, created_by, last_post_by, last_post_at, post_count
FROM forum_topics
WHERE handler = 'private';

-- name: AdminListGrantsByTopicID :many
SELECT g.idgrants, g.section, g.action, r.name AS role_name, u.username
FROM grants g
LEFT JOIN roles r ON g.role_id = r.idroles
LEFT JOIN users u ON g.user_id = u.idusers
WHERE g.section = 'privateforum' AND g.id = ?;

-- name: AdminListAllPrivateForumThreads :many
SELECT t.idforumthreads, t.idforumtopic, t.title, t.created_at, t.created_by, t.last_post_by, t.last_post_at, t.post_count, ft.title as topic_title
FROM forum_threads t
JOIN forum_topics ft ON t.idforumtopic = ft.idforumtopic
WHERE ft.handler = 'private';

-- name: AdminListGrantsByThreadID :many
SELECT g.idgrants, g.section, g.action, r.name AS role_name, u.username
FROM grants g
LEFT JOIN roles r ON g.role_id = r.idroles
LEFT JOIN users u ON g.user_id = u.idusers
WHERE g.section = 'privateforum_thread' AND g.id = ?;
