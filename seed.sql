-- Basic role definitions
INSERT INTO roles (name, can_login, is_admin) VALUES
  ('anonymous', 0, 0),
  ('user', 1, 0),
  ('content writer', 1, 0),
  ('moderator', 1, 0),
  ('administrator', 1, 1),
  ('rejected', 0, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

-- Grant user role to all users without any role
INSERT INTO user_roles (users_idusers, role_id)
SELECT u.idusers, r.id
FROM users u
JOIN roles r ON r.name='user'
LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
WHERE ur.iduser_roles IS NULL;

-- Role hierarchy grants
INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), r_admin.id, 'role', 'moderator', 1
FROM roles r_admin
WHERE r_admin.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), r_admin.id, 'role', 'content writer', 1
FROM roles r_admin
WHERE r_admin.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), r_mod.id, 'role', 'user', 1
FROM roles r_mod
WHERE r_mod.name = 'moderator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), r_writer.id, 'role', 'user', 1
FROM roles r_writer
WHERE r_writer.name = 'content writer'
ON DUPLICATE KEY UPDATE action=VALUES(action);
