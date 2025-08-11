ALTER TABLE roles
    ADD COLUMN private_labels TINYINT(1) NOT NULL DEFAULT 1;
UPDATE roles SET private_labels = can_login;

-- Seed labeler role and grants
INSERT INTO roles (name, can_login, is_admin)
SELECT 'labeler', 1, 0
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'labeler');

INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), r.id, g.section, 'label', 1
FROM roles r
JOIN (
    SELECT DISTINCT section FROM grants WHERE action IN ('see', 'view')
) g
WHERE r.name = 'labeler';

-- Grant label rights to all logged-in roles with view access
INSERT INTO grants (created_at, role_id, section, action, active)
SELECT NOW(), g.role_id, g.section, 'label', 1
FROM grants g
JOIN roles r ON r.id = g.role_id
WHERE g.action IN ('see', 'view')
  AND r.can_login = 1;

CREATE TABLE content_read_markers (
    id INT NOT NULL AUTO_INCREMENT,
    item VARCHAR(64) NOT NULL,
    item_id INT NOT NULL,
    user_id INT NOT NULL,
    last_comment_id INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY content_read_markers_uq (item, item_id, user_id)
);

-- Update schema version
UPDATE schema_version SET version = 68 WHERE version = 67;
