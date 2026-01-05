-- name: SystemGetForumTopicByID :one
SELECT idforumtopic, title, description, handler
FROM forumtopic
WHERE idforumtopic = ?;

-- name: SystemGetBlogEntryByID :one
SELECT idblogs, blog
FROM blogs
WHERE idblogs = ?;

-- name: SystemGetNewsPostByID :one
SELECT n.idsitenews, t.title AS topic_title
FROM sitenews n
JOIN forumthread th ON n.forumthread_id = th.idforumthread
JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
WHERE n.idsitenews = ?;

-- name: SystemGetWritingByID :one
SELECT idwriting, title, abstract
FROM writing
WHERE idwriting = ?;

-- name: SystemGetThreadByID :one
SELECT idforumthread, forumtopic_idforumtopic, lastposter
FROM forumthread
WHERE idforumthread = ?;
