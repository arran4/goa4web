-- Allow non-unique unverified emails
ALTER TABLE user_emails
    DROP INDEX user_emails_email_idx,
    ADD UNIQUE KEY user_emails_email_code_idx (email(255), last_verification_code);

-- Seed grants for linker permissions
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'category', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'category', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'see', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'view', 1
FROM roles r
WHERE r.name = 'anonymous'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'comment', 1
FROM roles r
WHERE r.name = 'user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'reply', 1
FROM roles r
WHERE r.name = 'user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'post', 1
FROM roles r
WHERE r.name IN ('content writer','administrator')
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', 'link', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'administrator'
ON DUPLICATE KEY UPDATE action=VALUES(action);

-- Update schema version
UPDATE schema_version SET version = 42 WHERE version = 41;
