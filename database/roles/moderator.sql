-- Role: moderator
-- Description: Moderator role with access to moderate content.
INSERT INTO roles (name, can_login, is_admin) VALUES ('moderator', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'lock', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'delete', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'move', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'imageboard', NULL, 'allow', 'approve', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'imageboard', NULL, 'allow', 'delete', 1
FROM roles r
WHERE r.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);
