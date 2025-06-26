-- name: RecalculateAllForumThreadMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    LIMIT 1
);

-- name: RecalculateForumThreadByIdMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_idforumthread = idforumthread
    LIMIT 1
)
WHERE idforumthread = ?;

-- name: GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions :one
SELECT th.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND th.idforumthread=?
ORDER BY t.lastaddition DESC;

-- name: MakeThread :execlastid
INSERT INTO forumthread (forumtopic_idforumtopic) VALUES (?);


-- name: GetForumTopicIdByThreadId :one
SELECT forumtopic_idforumtopic FROM forumthread WHERE idforumthread = ?;

-- name: DeleteForumThread :exec
UPDATE forumthread SET deleted_at = NOW() WHERE idforumthread = ?;


-- name: GetThreadsStartedByUser :many
SELECT th.*
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;
