-- Role: private forum user
-- Description: Grants access to private forums.
INSERT INTO roles (name, can_login, is_admin) VALUES ('private forum user', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'see', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'view', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'post', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'privateforum', NULL, 'allow', 'create', 1
FROM roles r
WHERE r.name = 'private forum user'
ON DUPLICATE KEY UPDATE action=VALUES(action);
