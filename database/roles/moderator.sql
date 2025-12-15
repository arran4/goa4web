-- Role: moderator
-- Description: Moderator role with access to moderate content.
INSERT INTO roles (name, can_login, is_admin) VALUES ('moderator', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_mod.id, 'role', NULL, 'allow', 'user', 1
FROM roles r_mod
WHERE r_mod.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);
