-- Seed grants for news permissions
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'comment', 1
FROM roles r
WHERE r.name = 'user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'post', 1
FROM roles r
WHERE r.name IN ('content writer','administrator')
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

UPDATE schema_version SET version = 38 WHERE version = 37;
