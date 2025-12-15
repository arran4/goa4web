-- Role: content writer
-- Description: Content writer role with access to create and manage content.
INSERT INTO roles (name, can_login, is_admin) VALUES ('content writer', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'see', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'view', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'post', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'see', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'view', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'post', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writings', NULL, 'allow', 'see', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writings', NULL, 'allow', 'view', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writings', NULL, 'allow', 'post', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writings', NULL, 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writings', NULL, 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);
