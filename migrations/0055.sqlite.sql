-- Migrate legacy 'Private discussion' category into private topics
ALTER TABLE forumtopic ADD COLUMN handler varchar(32) NOT NULL DEFAULT '';

UPDATE forumtopic
SET handler='private', forumcategory_idforumcategory = 0
WHERE forumcategory_idforumcategory IN (
    SELECT idforumcategory FROM forumcategory WHERE title='Private discussion'
);

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, 1, 'forum', 'topic', 'allow', t.idforumtopic, 'see', 1
FROM forumtopic t
WHERE t.handler='private'
  AND NOT EXISTS (
      SELECT 1 FROM grants g WHERE g.user_id=1 AND g.section='forum' AND g.item='topic' AND g.action='see' AND g.item_id=t.idforumtopic
  );

INSERT INTO grants (created_at, user_id, section, item, rule_type, item_id, action, active)
SELECT CURRENT_TIMESTAMP, 1, 'forum', 'topic', 'allow', t.idforumtopic, 'view', 1
FROM forumtopic t
WHERE t.handler='private'
  AND NOT EXISTS (
      SELECT 1 FROM grants g WHERE g.user_id=1 AND g.section='forum' AND g.item='topic' AND g.action='view' AND g.item_id=t.idforumtopic
  );

INSERT INTO grants (created_at, user_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, u.idusers, 'privateforum', 'topic', 'allow', 'see', 1
FROM users u
JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.deleted_at IS NULL
GROUP BY u.idusers;

INSERT INTO grants (created_at, user_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, u.idusers, 'privateforum', 'topic', 'allow', 'create', 1
FROM users u
JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.deleted_at IS NULL
GROUP BY u.idusers;

-- Grant login roles full access to the private forum
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'see', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'view', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'reply', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'post', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'edit', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'privateforum', 'topic', 'allow', 'create', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

DELETE FROM forumcategory WHERE title='Private discussion';

-- Update schema version
UPDATE schema_version SET version = 55 WHERE version = 54;
