-- +goose Up
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'images', 'upload', 'allow', 'see', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);