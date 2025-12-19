-- name: AdminListAllPrivateTopics :many
SELECT
    idforumtopic,
    COALESCE(title, '') AS title,
    lastposter AS last_post_by,
    lastaddition AS last_post_at,
    comments AS post_count
FROM
    forumtopic
WHERE
    handler = 'private';

-- name: AdminListGrantsByTopicID :many
SELECT
    g.id,
    g.section,
    g.action,
    r.name AS role_name,
    u.username
FROM
    grants g
LEFT JOIN
    roles r ON g.role_id = r.id
LEFT JOIN
    users u ON g.user_id = u.idusers
WHERE
    g.section = 'privateforum' AND g.item_id = ?;

-- name: AdminListAllPrivateForumThreads :many
SELECT
    t.idforumthread,
    t.forumtopic_idforumtopic as idforumtopic,
    COALESCE(SUBSTRING(c.text, 1, 100), 'unknown') AS title,
    c.written as created_at,
    c.users_idusers as created_by,
    t.lastposter as last_post_by,
    t.lastaddition as last_post_at,
    t.comments as post_count,
    COALESCE(ft.title, '') as topic_title
FROM
    forumthread t
JOIN
    forumtopic ft ON t.forumtopic_idforumtopic = ft.idforumtopic
JOIN
    comments c ON t.firstpost = c.idcomments
WHERE
    ft.handler = 'private';

-- name: AdminListGrantsByThreadID :many
SELECT
    g.id,
    g.section,
    g.action,
    r.name AS role_name,
    u.username
FROM
    grants g
LEFT JOIN
    roles r ON g.role_id = r.id
LEFT JOIN
    users u ON g.user_id = u.idusers
WHERE
    g.section = 'privateforum_thread' AND g.item_id = ?;
