-- name: CopyUserToDeactivated :exec
INSERT INTO deactivated_users (idusers, email, passwd, passwd_algorithm, username)
SELECT u.idusers, u.email, u.passwd, u.passwd_algorithm, u.username FROM users u WHERE u.idusers = ?;

-- name: AnonymizeUser :exec
UPDATE users
SET email = CONCAT('scrubbed_', idusers, '@example.com'),
    passwd = '',
    passwd_algorithm = '',
    username = CONCAT('scrubbed_', idusers)
WHERE idusers = ?;

-- name: RemoveUserByID :exec
DELETE FROM users WHERE idusers = ?;

-- name: CopyCommentsToDeactivatedByUser :exec
INSERT INTO deactivated_comments (idcomments, forumthread_idforumthread, users_idusers, language_idlanguage, written, text)
SELECT c.idcomments, c.forumthread_idforumthread, c.users_idusers, c.language_idlanguage, c.written, c.text FROM comments c WHERE c.users_idusers = ?;

-- name: AnonymizeCommentsByUser :exec
UPDATE comments
SET text = 'lorem ipsum dolor sit amet'
WHERE users_idusers = ?;

-- name: RemoveCommentsByUser :exec
DELETE FROM comments WHERE users_idusers = ?;

-- name: CopyUserFromDeactivated :exec
UPDATE users u
JOIN deactivated_users du ON u.idusers = du.idusers
SET u.email = du.email,
    u.passwd = du.passwd,
    u.passwd_algorithm = du.passwd_algorithm,
    u.username = du.username
WHERE u.idusers = ?;

-- name: RemoveDeactivatedUserByID :exec
DELETE FROM deactivated_users WHERE idusers = ?;

-- name: CopyCommentsFromDeactivatedByUser :exec
UPDATE comments c
JOIN deactivated_comments dc ON c.idcomments = dc.idcomments
SET c.text = dc.text
WHERE c.users_idusers = ?;

-- name: RemoveDeactivatedCommentsByUser :exec
DELETE FROM deactivated_comments WHERE users_idusers = ?;

-- name: DeletePermissionsByUser :exec
DELETE FROM permissions WHERE users_idusers = ?;

-- name: DeleteSessionsByUser :exec
DELETE FROM sessions WHERE users_idusers = ?;

-- name: GetDeactivatedUserByUsername :one
SELECT * FROM deactivated_users WHERE username = ?;
