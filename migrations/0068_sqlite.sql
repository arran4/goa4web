-- +goose Up
ALTER TABLE roles
    ADD COLUMN private_labels TINYINT(1) NOT NULL DEFAULT 1;

UPDATE roles SET private_labels = can_login;

INSERT INTO grants (created_at, role_id, section, action, active, rule_type)
SELECT CURRENT_TIMESTAMP, g.role_id, g.section, 'label', 1, 'allow'
FROM grants g
         JOIN roles r ON r.id = g.role_id
WHERE g.action IN ('see', 'view')
  AND r.can_login = 1;

CREATE TABLE IF NOT EXISTS content_read_markers (
id INTEGER PRIMARY KEY AUTOINCREMENT,
item TEXT NOT NULL,
item_id INT NOT NULL,
user_id INT NOT NULL,
last_comment_id INT NOT NULL,
UNIQUE (item, item_id, user_id)
);
UPDATE schema_version SET version = 68;