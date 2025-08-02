-- Grant upload listing to login roles
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'images', 'upload', 'allow', 'see', 1
FROM roles r
WHERE r.can_login = 1
ON CONFLICT DO NOTHING;

-- Update schema version
UPDATE schema_version SET version = 52 WHERE version = 51;
