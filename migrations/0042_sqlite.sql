-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS user_emails_email_code_idx ON user_emails (email, last_verification_code);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'category', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'category', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'comment', 1
FROM roles r
WHERE r.name = 'user';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'user';

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'post', 1
FROM roles r
WHERE r.name IN ('content writer','administrator');

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', 'link', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'administrator';

UPDATE schema_version SET version = 42;