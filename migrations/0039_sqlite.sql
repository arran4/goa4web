-- +goose Up
INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'see', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 2;

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'view', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 2;

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, utp.users_idusers, 'forum', 'topic', 'allow', utp.forumtopic_idforumtopic, 'edit', 1
FROM user_topic_permissions utp
WHERE utp.role_id = 4;

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, tp.see_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'see', 1
FROM topic_permissions tp
WHERE tp.see_role_id IS NOT NULL;

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, tp.view_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'view', 1
FROM topic_permissions tp
WHERE tp.view_role_id IS NOT NULL;

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, tp.reply_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'reply', 1
FROM topic_permissions tp
WHERE tp.reply_role_id IS NOT NULL;

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, tp.newthread_role_id, 'forum', 'topic', 'allow', tp.forumtopic_idforumtopic, 'post', 1
FROM topic_permissions tp
WHERE tp.newthread_role_id IS NOT NULL;

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'see', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'anonymous'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'view', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'anonymous'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'reply', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'user'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic);

INSERT INTO grants (created_at, role_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'forum', 'topic', 'allow', t.idforumtopic, 'post', 1
FROM forumtopic t
CROSS JOIN roles r
WHERE r.name = 'user'
  AND NOT EXISTS (SELECT 1 FROM topic_permissions tp WHERE tp.forumtopic_idforumtopic = t.idforumtopic);

UPDATE schema_version SET version = 39;