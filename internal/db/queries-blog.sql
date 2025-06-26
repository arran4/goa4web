-- name: UpdateBlogEntry :exec
UPDATE blogs
SET language_idlanguage = ?, blog = ?
WHERE idblogs = ?;

-- name: CreateBlogEntry :execlastid
INSERT INTO blogs (users_idusers, language_idlanguage, blog, written)
VALUES (?, ?, ?, NOW());

-- name: AssignThreadIdToBlogEntry :exec
UPDATE blogs
SET forumthread_idforumthread = ?
WHERE idblogs = ?;

-- name: GetBlogEntriesForUserDescending :many
SELECT b.idblogs, b.forumthread_idforumthread, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE (b.language_idlanguage = sqlc.arg(Language_idlanguage) OR sqlc.arg(Language_idlanguage) = 0)
AND (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogEntriesForUserDescendingLanguages :many
SELECT b.idblogs, b.forumthread_idforumthread, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE (b.users_idusers = sqlc.arg(Users_idusers) OR sqlc.arg(Users_idusers) = 0)
AND (
    NOT EXISTS (SELECT 1 FROM userlang ul WHERE ul.users_idusers = sqlc.arg(Viewer_idusers))
    OR b.language_idlanguage IN (
        SELECT ul.language_idlanguage FROM userlang ul WHERE ul.users_idusers = sqlc.arg(Viewer_idusers)
    )
)
ORDER BY b.written DESC
LIMIT ? OFFSET ?;

-- name: GetBlogEntriesByIdsDescending :many
SELECT b.idblogs, b.forumthread_idforumthread, b.users_idusers, b.language_idlanguage, b.blog, b.written
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.idblogs IN (sqlc.slice(blogIds))
ORDER BY b.written DESC
;

-- name: GetBlogEntryForUserById :one
SELECT b.idblogs, b.forumthread_idforumthread, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.idblogs = ?
LIMIT 1;

-- name: GetCountOfBlogPostsByUser :many
SELECT u.username, COUNT(b.idblogs)
FROM blogs b, users u
WHERE b.users_idusers = u.idusers
GROUP BY u.idusers;

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


-- name: GetAllBlogEntriesByUser :many
SELECT b.idblogs, b.forumthread_idforumthread, b.users_idusers, b.language_idlanguage, b.blog, b.written, u.username, coalesce(th.comments, 0)
FROM blogs b
LEFT JOIN users u ON b.users_idusers=u.idusers
LEFT JOIN forumthread th ON b.forumthread_idforumthread = th.idforumthread
WHERE b.users_idusers = ?
ORDER BY b.written DESC;
