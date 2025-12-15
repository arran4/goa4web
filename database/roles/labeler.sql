-- Role: labeler
-- Description: Role for users who can label content.
INSERT INTO roles (name, can_login, is_admin) VALUES ('labeler', 1, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r_labeler.id, g.section, NULL, 'allow', 'label', 1
FROM roles r_labeler
JOIN (
    SELECT DISTINCT section FROM grants WHERE action IN ('see', 'view')
) g
WHERE r_labeler.name = 'labeler'
ON DUPLICATE KEY UPDATE action=VALUES(action);
