-- Convert user_topic_permissions to grants and drop old tables
INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'see', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 2
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'view', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 2
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'edit', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 4
ON DUPLICATE KEY UPDATE action=VALUES(action);

-- Convert topic permissions to grants
INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), tp.see_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'see', 1
FROM topic_permissions tp
WHERE tp.see_role_id IS NOT NULL
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), tp.view_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'view', 1
FROM topic_permissions tp
WHERE tp.view_role_id IS NOT NULL
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), tp.reply_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'reply', 1
FROM topic_permissions tp
WHERE tp.reply_role_id IS NOT NULL
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), tp.newthread_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'post', 1
FROM topic_permissions tp
WHERE tp.newthread_role_id IS NOT NULL
ON DUPLICATE KEY UPDATE action=VALUES(action);

-- Default grants for public topics
INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'see', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'anonymous'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic)
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'view', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'anonymous'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic)
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'reply', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name IN ('user','content writer','moderator','administrator')
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic)
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'post', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name IN ('user','content writer','moderator','administrator')
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic)
ON DUPLICATE KEY UPDATE action=VALUES(action);

DROP TABLE user_topic_permissions;
DROP TABLE topic_permissions;

-- Update schema version
UPDATE schema_version SET version = 39 WHERE version = 38;
