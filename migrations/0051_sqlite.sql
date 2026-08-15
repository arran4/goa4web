-- +goose Up
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'search', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'news', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'forum', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'linker', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'blogs', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT NOW(), r.id, 'writing', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

CREATE TABLE IF NOT EXISTS external_links (
id INTEGER PRIMARY KEY AUTOINCREMENT,
url TEXT NOT NULL,
clicks INT NOT NULL DEFAULT 0,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
updated_by INT DEFAULT NULL,
card_title TEXT,
card_description TEXT,
card_image TEXT,
card_image_cache TEXT,
favicon_cache TEXT,
UNIQUE (url)
);