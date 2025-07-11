-- Basic role definitions
INSERT INTO roles (name) VALUES
  ('anonymous'),
  ('normal user'),
  ('content writer'),
  ('moderator'),
  ('administrator')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Grant normal user role to all users without any role
INSERT INTO user_roles (users_idusers, role_id)
SELECT u.idusers, r.id
FROM users u
JOIN roles r ON r.name='normal user'
LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
WHERE ur.iduser_roles IS NULL;
