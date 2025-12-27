-- Role: faq reader
-- Description: FAQ reader role with access to read FAQs.
INSERT INTO roles (name, can_login, is_admin) VALUES ('faq reader', 0, 0)
ON DUPLICATE KEY UPDATE name = VALUES(name), can_login = VALUES(can_login), is_admin = VALUES(is_admin);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'faq', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.name = 'faq reader'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'faq', 'question/answer', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'faq reader'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'faq', 'question/answer', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'faq reader'
ON DUPLICATE KEY UPDATE action=VALUES(action);
