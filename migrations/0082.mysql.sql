INSERT INTO grants (
    created_at, role_id, section, item, rule_type, action, active
)
SELECT NOW(), r.id, 'forum', 'comment', 'allow', 'edit', 1
FROM roles r
WHERE r.name = 'user'
ON DUPLICATE KEY UPDATE action=VALUES(action);

UPDATE schema_version SET version = 82;
