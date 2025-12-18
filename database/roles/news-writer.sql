-- Role: news writer
-- Description: News writer role with access to create and edit news.
INSERT INTO roles (name, can_login, is_admin) VALUES ('news writer', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'post', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', 'post', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'news writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);
