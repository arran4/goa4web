-- name: SystemGetForumReplyNotificationSource :one
SELECT topic.idforumtopic AS topic_id, topic.title AS topic_title, topic.handler,
       thread.idforumthread AS thread_id, reply.text AS reply_text, author.username
FROM comments reply
JOIN forumthread thread ON thread.idforumthread = reply.forumthread_id
JOIN forumtopic topic ON topic.idforumtopic = thread.forumtopic_idforumtopic
JOIN comments opener ON opener.idcomments = thread.firstpost AND opener.forumthread_id = thread.idforumthread
LEFT JOIN users author ON author.idusers = reply.users_idusers AND author.deleted_at IS NULL
WHERE reply.idcomments = sqlc.arg(comment_id)
  AND reply.users_idusers = sqlc.arg(actor_id)
  AND reply.deleted_at IS NULL AND opener.deleted_at IS NULL
  AND thread.deleted_at IS NULL AND topic.deleted_at IS NULL
  AND (sqlc.arg(append_id) = 0 OR EXISTS (
      SELECT 1 FROM comment_appends activity
      WHERE activity.id = sqlc.arg(append_id)
        AND activity.comment_id = reply.idcomments
        AND activity.users_idusers = reply.users_idusers
  ));
