-- Seed search permissions for all sections
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'search', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writing', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

-- Update schema version
UPDATE schema_version SET version = 51 WHERE version = 50;
