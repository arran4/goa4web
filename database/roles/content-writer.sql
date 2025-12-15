-- Role: content writer
-- Description: Content writer role with access to create and manage content.
INSERT INTO roles (name, can_login, is_admin) VALUES ('content writer', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_writer.id, 'role', NULL, 'allow', 'user', 1
FROM roles r_writer
WHERE r_writer.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);
