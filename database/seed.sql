-- Basic role definitions
INSERT INTO roles (name, can_login, is_admin) VALUES
  ('anyone', 0, 0),
  ('user', 1, 0),
  ('content writer', 1, 0),
  ('moderator', 1, 0),
  ('labeler', 1, 0),
  ('administrator', 1, 1),
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

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_mod.id, 'role', NULL, 'allow', 'user', 1
FROM roles r_mod
WHERE r_mod.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_writer.id, 'role', NULL, 'allow', 'user', 1
FROM roles r_writer
WHERE r_writer.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_labeler.id, g.section, NULL, 'allow', 'label', 1
FROM roles r_labeler
JOIN (
    SELECT DISTINCT section FROM grants WHERE action IN ('see', 'view')
) g
WHERE r_labeler.name = 'labeler'
ON DUPLICATE KEY UPDATE action=VALUES(action);

-- Grant label rights to all logged-in roles with view access
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), g.role_id, g.section, NULL, 'allow', 'label', 1
FROM grants g
JOIN roles r ON r.id = g.role_id
WHERE g.action IN ('see', 'view')
  AND r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
