-- Migrate legacy 'Private discussion' category into private topics
ALTER TABLE forumtopic ADD COLUMN handler varchar(32) NOT NULL DEFAULT '';

UPDATE forumtopic t
JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
SET t.handler='private', t.forumcategory_idforumcategory = 0
WHERE fc.title='Private discussion';

-- Grant participants of existing private topics the new privateforum grants
INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), g.user_id, 'privateforum', 'topic', 'allow', g.item_id, 'see', 1
FROM grants g
JOIN forumtopic t ON g.item_id = t.idforumtopic
WHERE t.handler = 'private' AND g.section = 'forum' AND g.item = 'topic' AND g.action = 'see'
  AND NOT EXISTS (
    SELECT 1 FROM grants g2 WHERE g2.user_id = g.user_id AND g2.section = 'privateforum' AND g2.item = 'topic' AND g2.action = 'see' AND g2.item_id = g.item_id
  );

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), g.user_id, 'privateforum', 'topic', 'allow', g.item_id, 'view', 1
FROM grants g
JOIN forumtopic t ON g.item_id = t.idforumtopic
WHERE t.handler = 'private' AND g.section = 'forum' AND g.item = 'topic' AND g.action = 'view'
  AND NOT EXISTS (
    SELECT 1 FROM grants g2 WHERE g2.user_id = g.user_id AND g2.section = 'privateforum' AND g2.item = 'topic' AND g2.action = 'view' AND g2.item_id = g.item_id
  );

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), g.user_id, 'privateforum', 'topic', 'allow', g.item_id, 'post', 1
FROM grants g
JOIN forumtopic t ON g.item_id = t.idforumtopic
WHERE t.handler = 'private' AND g.section = 'forum' AND g.item = 'topic' AND g.action = 'post'
    AND NOT EXISTS (
        SELECT 1 FROM grants g2 WHERE g2.user_id = g.user_id AND g2.section = 'privateforum' AND g2.item = 'topic' AND g2.action = 'post' AND g2.item_id = g.item_id
    );

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), g.user_id, 'privateforum', 'topic', 'allow', g.item_id, 'reply', 1
FROM grants g
JOIN forumtopic t ON g.item_id = t.idforumtopic
WHERE t.handler = 'private' AND g.section = 'forum' AND g.item = 'topic' AND g.action = 'reply'
    AND NOT EXISTS (
        SELECT 1 FROM grants g2 WHERE g2.user_id = g.user_id AND g2.section = 'privateforum' AND g2.item = 'topic' AND g2.action = 'reply' AND g2.item_id = g.item_id
    );

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT NOW(), g.user_id, 'privateforum', 'topic', 'allow', g.item_id, 'edit', 1
FROM grants g
JOIN forumtopic t ON g.item_id = t.idforumtopic
WHERE t.handler = 'private' AND g.section = 'forum' AND g.item = 'topic'
    AND NOT EXISTS (
        SELECT 1 FROM grants g2 WHERE g2.user_id = g.user_id AND g2.section = 'privateforum' AND g2.item = 'topic' AND g2.action = 'edit' AND g2.item_id = g.item_id
    );

INSERT INTO grants (created_at, user_id, section, item, rule_type, action, active)
SELECT NOW(), u.idusers, 'privateforum', 'topic', 'allow', 'see', 1
FROM users u
JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.deleted_at IS NULL
GROUP BY u.idusers;

INSERT INTO grants (created_at, user_id, section, item, rule_type, action, active)
SELECT NOW(), u.idusers, 'privateforum', 'topic', 'allow', 'create', 1
FROM users u
JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.deleted_at IS NULL
GROUP BY u.idusers;

-- Grant login roles full access to the private forum
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'see', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'view', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'reply', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'post', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'edit', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', 'topic', 'allow', 'create', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

DELETE FROM forumcategory WHERE title='Private discussion';

-- Update schema version
UPDATE schema_version SET version = 55 WHERE version = 54;
