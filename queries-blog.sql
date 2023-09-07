-- name: Update_blog :exec
UPDATE blogs
SET language_idlanguage = ?, blog = ?
WHERE idblogs = ?;

-- name: Add_blog :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
VALUES (?, ?, ?, NOW());

-- name: Assign_blog_to_thread :exec
UPDATE blogs
SET forumthread_idforumthread = ?
WHERE idblogs = ?;

-- name: Show_latest_blogs :many
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogs :many
SELECT b.*
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
-- WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
-- AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
-- AND b.idblogs IN (sqlc.slice(blogIds))
WHERE b.idblogs IN (sqlc.slice(blogIds))
ORDER BY b.written DESC
-- LIMIT ? OFFSET ?
;

-- name: Show_blog :one
SELECT b.blog, b.written, u.username, b.idblogs, coalesce(th.comments, 0), b.users_idusers, b.forumthread_idforumthread
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.idblogs = ?
LIMIT 1;

-- name: Show_blogger_list :many
SELECT u.username, COUNT(b.idblogs)
FROM blogs b, users u
WHERE b.users_idusers = u.idusers
GROUP BY u.idusers;

-- name: Blog_atom :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers = ?
ORDER BY b.written DESC
LIMIT ?;

-- name: Blog_rss :many
SELECT b.idblogs, LEFT(b.written, 255), b.blog, u.username
FROM blogs b, users u
WHERE u.idusers = b.users_idusers AND b.users_idusers= ?
ORDER BY b.written DESC
LIMIT ?;

-- name: Show_blog_edit :one
SELECT b.blog, b.language_idlanguage
FROM blogs b, users u
WHERE b.users_idusers = u.idusers AND b.idblogs = ?
LIMIT 1;

-- name: BlogsSearchFirst :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
;

-- name: BlogsSearchNext :many
SELECT DISTINCT cs.blogs_idblogs
FROM blogsSearch cs
LEFT JOIN searchwordlist swl ON swl.idsearchwordlist=cs.searchwordlist_idsearchwordlist
WHERE swl.word=?
AND cs.blogs_idblogs IN (sqlc.slice('ids'))
;

