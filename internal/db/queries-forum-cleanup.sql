-- name: DeleteThreadsByTopicID :exec
DELETE forumthread, comments, comments_search
FROM forumthread
LEFT JOIN comments ON comments.forumthread_id = forumthread.idforumthread
LEFT JOIN comments_search ON comments_search.comment_id = comments.idcomments
WHERE forumthread.forumtopic_idforumtopic = ?;

-- name: AdminListOrphanForumThreads :many
SELECT t.idforumthread
FROM forumthread t
LEFT JOIN forumtopic ft ON t.forumtopic_idforumtopic = ft.idforumtopic
WHERE ft.idforumtopic IS NULL;

-- name: AdminListOrphanComments :many
SELECT c.idcomments
FROM comments c
LEFT JOIN forumthread t ON c.forumthread_id = t.idforumthread
WHERE t.idforumthread IS NULL;
