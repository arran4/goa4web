-- +goose Up
INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'search', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'news', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'forum', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'linker', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'blogs', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

INSERT INTO grants (created_at, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, r.id, 'writing', NULL, 'allow', 'search', 1
FROM roles r
WHERE r.can_login = 1;

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
UPDATE schema_version SET version = 51;