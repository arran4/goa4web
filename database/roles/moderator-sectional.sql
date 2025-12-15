-- Role: moderator-sectional
-- Description: Moderator role with access to moderate content in a specific section.
INSERT INTO roles (name, can_login, is_admin) VALUES ('moderator-sectional', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);
