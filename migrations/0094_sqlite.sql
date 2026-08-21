-- +goose Up
UPDATE grants
SET section = 'privateforum_thread'
WHERE section = 'privateforum'
  AND item = 'thread'
  AND rule_type = 'allow'
  AND active = 1
  AND item_id IS NOT NULL
  AND action IN ('view', 'reply')
  AND (user_id IS NOT NULL OR role_id IS NOT NULL);

INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type,
    item_id, item_rule, action, extra, active
)
SELECT DISTINCT
    CURRENT_TIMESTAMP, topic_grant.user_id, topic_grant.role_id,
    'privateforum_thread', 'thread', 'allow',
    thread_row.idforumthread, NULL, topic_grant.action, NULL, 1
FROM grants topic_grant
JOIN forumtopic topic
    ON topic.idforumtopic = topic_grant.item_id
   AND topic.handler = 'private'
JOIN forumthread thread_row
    ON thread_row.forumtopic_idforumtopic = topic.idforumtopic
WHERE topic_grant.section = 'privateforum'
  AND topic_grant.item = 'topic'
  AND topic_grant.rule_type = 'allow'
  AND topic_grant.active = 1
  AND topic_grant.action IN ('view', 'reply')
  AND (topic_grant.user_id IS NOT NULL OR topic_grant.role_id IS NOT NULL)
  AND NOT EXISTS (
      SELECT 1
      FROM grants thread_grant
        WHERE thread_grant.section = 'privateforum_thread'
          AND thread_grant.item = 'thread'
          AND thread_grant.rule_type = 'allow'
          AND thread_grant.item_id = thread_row.idforumthread
        AND thread_grant.action = topic_grant.action
        AND thread_grant.active = 1
        AND (thread_grant.user_id IS topic_grant.user_id)
        AND (thread_grant.role_id IS topic_grant.role_id)
  );

UPDATE schema_version SET version = 94;