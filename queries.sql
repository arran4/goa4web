-- name: ListUsersSubscribedToBlogs :many
SELECT u.email FROM blogs t, users u, preferences p
WHERE t.idblogs=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToLinker :many
SELECT u.email FROM linker t, users u, preferences p
WHERE t.idlinker=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToWriting :many
SELECT u.email FROM writing t, users u, preferences p
WHERE t.idwriting=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToThread :many
SELECT u.email FROM comments c, users u, preferences p
WHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

