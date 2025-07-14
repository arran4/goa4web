-- Basic role definitions
INSERT INTO roles (name) VALUES
  ('anonymous'),
  ('user'),
  ('content writer'),
  ('moderator'),
  ('administrator')
ON DUPLICATE KEY UPDATE name = VALUES(name);

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
