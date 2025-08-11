ALTER TABLE roles
    ADD COLUMN private_labels BOOLEAN NOT NULL DEFAULT TRUE;
UPDATE roles SET private_labels = can_login;

-- Grant label rights to all logged-in roles with view access
INSERT INTO grants (created_at, role_id, section, action, active, rule_type)
SELECT NOW(), g.role_id, g.section, 'label', 1, 'allow'
FROM grants g
         JOIN roles r ON r.id = g.role_id
WHERE g.action IN ('see', 'view')
  AND r.can_login = TRUE;

CREATE TABLE content_read_markers (
    id SERIAL PRIMARY KEY,
    item TEXT NOT NULL,
    item_id INT NOT NULL,
    user_id INT NOT NULL,
    last_comment_id INT NOT NULL,
    UNIQUE (item, item_id, user_id)
);

-- Update schema version
UPDATE schema_version SET version = 68 WHERE version = 67;
