-- name: UpdateForumCategory :exec
UPDATE forumcategory SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumcategory = ?;

-- name: GetAllForumCategoriesWithSubcategoryCount :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount,
       COUNT(t.idforumtopic)   as TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
GROUP BY c.idforumcategory;

-- name: GetAllForumTopics :many
SELECT t.*
FROM forumtopic t
GROUP BY t.idforumtopic;

-- name: UpdateForumTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumtopic = ?;

-- name: GetAllForumTopicsByCategoryIdForUserWithLastPosterName :many
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = ? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: GetAllForumTopicsForUser :many
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY t.lastaddition DESC;

-- name: GetForumTopicByIdForUser :one
SELECT t.*, lu.username AS LastPosterUsername, r.seelevel, u.level
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0) AND t.idforumtopic=?
ORDER BY t.lastaddition DESC;

-- name: DeleteUsersForumTopicLevelPermission :exec
DELETE FROM userstopiclevel WHERE forumtopic_idforumtopic = ? AND users_idusers = ?;

-- name: UpsertUsersForumTopicLevelPermission :exec
INSERT INTO userstopiclevel (forumtopic_idforumtopic, users_idusers, level, invitemax)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE level = VALUES(level), invitemax = VALUES(invitemax);

-- name: GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopic :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic
WHERE u.idusers = ?;

-- name: GetAllForumTopicsWithPermissionsAndTopic :many
SELECT u.*, t.*, utl.*, tr.*
FROM users u
JOIN userstopiclevel utl ON utl.users_idusers=u.idusers
JOIN forumtopic t ON utl.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN topicrestrictions tr ON t.idforumtopic = tr.forumtopic_idforumtopic;

-- name: GetAllForumCategories :many
SELECT f.*
FROM forumcategory f;

-- name: CreateForumCategory :exec
INSERT INTO forumcategory (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: CreateForumTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: FindForumTopicByTitle :one
SELECT *
FROM forumtopic
WHERE title=?;

-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many
SELECT th.*, lu.username AS lastposterusername, lu.idusers AS lastposterid, fcu.username as firstpostusername, fc.written as firstpostwritten, fc.text as firstposttext
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
LEFT JOIN userstopiclevel u ON u.forumtopic_idforumtopic = t.idforumtopic AND u.users_idusers = ?
LEFT JOIN users lu ON lu.idusers = t.lastposter
LEFT JOIN comments fc ON th.firstpost=fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE th.forumtopic_idforumtopic=? AND IF(r.seelevel IS NOT NULL, r.seelevel , 0) <= IF(u.level IS NOT NULL, u.level, 0)
ORDER BY th.lastaddition DESC;

-- name: RebuildAllForumTopicMetaColumns :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
);

-- name: RebuildForumTopicByIdMetaColumns :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
)
WHERE idforumtopic = ?;

-- name: DeleteForumCategory :exec
DELETE FROM forumcategory WHERE idforumcategory = ?;

-- name: DeleteForumTopic :exec
-- Removes a forum topic by ID.
DELETE FROM forumtopic WHERE idforumtopic = ?;


