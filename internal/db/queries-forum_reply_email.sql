-- name: SystemQueueForumReplyEmailActivity :exec
INSERT INTO forum_reply_email_activities
    (recipient_id, activity_position, thread_id, window_end, preference_generation, attempted_at)
SELECT u.idusers, sqlc.arg(activity_position), sqlc.arg(thread_id), sqlc.arg(window_end),
       COALESCE(p.optional_email_generation, 0),
       CASE WHEN COALESCE(p.emailforumupdates, 0) = 0 THEN 0 ELSE NULL END
FROM users u
LEFT JOIN preferences p ON p.idpreferences = (
    SELECT p2.idpreferences FROM preferences p2 WHERE p2.users_idusers = u.idusers
    ORDER BY p2.idpreferences DESC LIMIT 1
)
WHERE u.idusers = sqlc.arg(recipient_id) AND u.deleted_at IS NULL
ON DUPLICATE KEY UPDATE activity_position = forum_reply_email_activities.activity_position;

-- name: SystemGetDueForumReplyEmailWindow :one
SELECT a.recipient_id, a.window_end
FROM forum_reply_email_activities a
LEFT JOIN forum_reply_email_activities recent ON recent.recipient_id = a.recipient_id
    AND (recent.lease_until > sqlc.arg(lease_available_at) OR recent.attempted_at > sqlc.arg(throttle_after))
WHERE recent.recipient_id IS NULL AND a.attempted_at IS NULL AND a.claim_count < 5
  AND a.window_end <= sqlc.arg(due_at) AND a.lease_until <= sqlc.arg(lease_available_at)
ORDER BY a.window_end, a.recipient_id, a.activity_position
LIMIT 1;

-- name: SystemClaimForumReplyEmailWindow :execrows
UPDATE forum_reply_email_activities
SET lease_token = sqlc.arg(lease_token), lease_until = sqlc.arg(lease_until), claim_count = claim_count + 1
WHERE recipient_id = sqlc.arg(recipient_id) AND window_end = sqlc.arg(window_end)
  AND attempted_at IS NULL AND claim_count < 5 AND lease_until <= sqlc.arg(lease_available_at);

-- name: SystemListEligibleForumReplyEmailThreads :many
WITH eligible AS (
    SELECT a.recipient_id, a.thread_id, th.forumtopic_idforumtopic AS topic_id, t.handler, t.title
    FROM forum_reply_email_activities a
    JOIN users u ON u.idusers = a.recipient_id AND u.deleted_at IS NULL
    JOIN preferences p ON p.idpreferences = (
        SELECT p2.idpreferences FROM preferences p2 WHERE p2.users_idusers = a.recipient_id
        ORDER BY p2.idpreferences DESC LIMIT 1
    ) AND p.emailforumupdates = 1 AND p.optional_email_generation = a.preference_generation
    JOIN forumthread th ON th.idforumthread = a.thread_id
    JOIN forumtopic t ON t.idforumtopic = th.forumtopic_idforumtopic
    JOIN comments opener ON opener.idcomments = th.firstpost
    JOIN comments reply ON reply.idcomments = (a.activity_position >> 32) AND reply.forumthread_id = a.thread_id
    WHERE a.recipient_id = sqlc.arg(recipient_id) AND a.window_end = sqlc.arg(window_end)
      AND a.lease_token = sqlc.arg(lease_token) AND a.attempted_at IS NULL
      AND reply.deleted_at IS NULL
      AND (
          NOT EXISTS (SELECT 1 FROM user_language ul WHERE ul.users_idusers = a.recipient_id)
          OR (
              (COALESCE(t.language_id, 0) = 0 OR t.language_id IN (SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = a.recipient_id))
              AND (COALESCE(opener.language_id, 0) = 0 OR opener.language_id IN (SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = a.recipient_id))
              AND (COALESCE(reply.language_id, 0) = 0 OR reply.language_id IN (SELECT ul.language_id FROM user_language ul WHERE ul.users_idusers = a.recipient_id))
          )
      )
      AND EXISTS (
          SELECT 1 FROM grants g
          WHERE g.section = CASE WHEN t.handler = 'private' THEN 'privateforum' ELSE 'forum' END
            AND (g.item = 'topic' OR g.item IS NULL) AND g.action = 'view' AND g.active = 1
            AND (g.item_id = t.idforumtopic OR (t.handler <> 'private' AND g.item_id IS NULL))
            AND (g.user_id = a.recipient_id OR g.user_id IS NULL)
            AND (g.role_id IS NULL OR g.role_id IN (SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = a.recipient_id)
                OR g.role_id IN (SELECT r.id FROM roles r WHERE r.name = 'anyone'))
      )
      AND (t.handler <> 'private' OR EXISTS (
          SELECT 1 FROM grants g
          WHERE g.section = 'privateforum_thread' AND g.item = 'thread' AND g.action = 'view'
            AND g.active = 1 AND g.item_id = a.thread_id
            AND (g.user_id = a.recipient_id OR g.user_id IS NULL)
            AND (g.role_id IS NULL OR g.role_id IN (SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = a.recipient_id)
                OR g.role_id IN (SELECT r.id FROM roles r WHERE r.name = 'anyone'))
      ))
)
SELECT e.thread_id, e.topic_id, e.handler, e.title, COUNT(*) AS reply_count
FROM eligible e
GROUP BY e.thread_id, e.topic_id, e.handler, e.title
ORDER BY e.thread_id
LIMIT ? OFFSET ?;

-- name: SystemSealForumReplyEmailWindow :execrows
UPDATE forum_reply_email_activities
SET attempted_at = sqlc.arg(attempted_at), lease_until = 0
WHERE recipient_id = sqlc.arg(recipient_id) AND window_end = sqlc.arg(window_end)
  AND lease_token = sqlc.arg(lease_token) AND attempted_at IS NULL AND lease_until > sqlc.arg(lease_valid_at);

-- name: SystemMarkForumReplyEmailWindowSent :exec
UPDATE forum_reply_email_activities
SET sent_at = sqlc.arg(sent_at)
WHERE recipient_id = sqlc.arg(recipient_id) AND window_end = sqlc.arg(window_end)
  AND lease_token = sqlc.arg(lease_token) AND attempted_at > 0;
