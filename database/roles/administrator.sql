-- Role: administrator
-- Description: Administrator role with full access.
INSERT INTO roles (name, can_login, is_admin) VALUES ('administrator', 1, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_admin.id, 'role', NULL, 'allow', 'moderator', 1
FROM roles r_admin
WHERE r_admin.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_admin.id, 'role', NULL, 'allow', 'content writer', 1
FROM roles r_admin
WHERE r_admin.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);
