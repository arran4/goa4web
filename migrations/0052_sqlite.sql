-- +goose Up
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'images', 'upload', 'allow', 'see', 1
FROM roles r
WHERE r.can_login = 1;
UPDATE schema_version SET version = 52;