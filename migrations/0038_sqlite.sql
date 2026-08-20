-- +goose Up
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'comment', 1
FROM roles r
WHERE r.name = 'user';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'user';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'post', 1
FROM roles r
WHERE r.name IN ('content writer','administrator');

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', 'post', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'administrator';

UPDATE schema_version SET version = 38;