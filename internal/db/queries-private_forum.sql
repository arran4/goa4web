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
    CAST(COALESCE(SUBSTRING(fp.text, 1, 100), 'unknown') AS CHAR) AS title,
    fp.written as created_at,
    fp.users_idusers as created_by,
    t.lastposter as last_post_by,
    t.lastaddition as last_post_at,
    t.comments as post_count,
    COALESCE(ft.title, '') as topic_title,
    CAST(COUNT(c.idcomments) AS SIGNED) AS total_comments,
    CAST(COALESCE(SUM(CASE WHEN c.text IS NOT NULL THEN 1 ELSE 0 END), 0) AS SIGNED) AS valid_comments,
    CAST(COALESCE(SUM(CASE WHEN c.text IS NULL THEN 1 ELSE 0 END), 0) AS SIGNED) AS invalid_comments
FROM
    forumthread t
JOIN
    forumtopic ft ON t.forumtopic_idforumtopic = ft.idforumtopic
LEFT JOIN
    comments fp ON t.firstpost = fp.idcomments
LEFT JOIN
    comments c ON c.forumthread_id = t.idforumthread
WHERE
    ft.handler = 'private'
GROUP BY
    t.idforumthread,
    t.forumtopic_idforumtopic,
    fp.text,
    fp.written,
    fp.users_idusers,
    t.lastposter,
    t.lastaddition,
    t.comments,
    ft.title;

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

-- name: AdminListPrivateForumInvalidCommentsByThread :many
SELECT
    idcomments
FROM
    comments
WHERE
    forumthread_id = ?
    AND text IS NULL;
