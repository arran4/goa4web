-- Basic role definitions
INSERT INTO roles (name, can_login, is_admin) VALUES
  ('anyone', 0, 0),
  ('user', 1, 0),
  ('rejected', 0, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

-- Ensure private label flag mirrors login capability
UPDATE roles SET private_labels = can_login;

-- Grant user role to all users without any role
INSERT INTO user_roles (users_idusers, role_id)
SELECT u.idusers, r.id
FROM users u
JOIN roles r ON r.name='user'
LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
WHERE ur.iduser_roles IS NULL;

-- Grant label rights to all logged-in roles with view access
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), g.role_id, g.section, NULL, 'allow', 'label', 1
FROM grants g
JOIN roles r ON r.id = g.role_id
WHERE g.action IN ('see', 'view')
  AND r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
