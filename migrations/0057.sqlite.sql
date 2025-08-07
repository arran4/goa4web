-- Grant public view access to blogs, writings, news and FAQs
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'blogs', 'entry', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'blogs', 'entry', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'writing', 'category', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'writing', 'category', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'writing', 'article', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'writing', 'article', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'news', 'post', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'news', 'post', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'faq', 'category', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'faq', 'category', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'faq', 'question/answer', 'allow', 'see', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT NOW(), NULL, NULL, 'faq', 'question/answer', 'allow', 'view', 1
ON DUPLICATE KEY UPDATE action=VALUES(action);

UPDATE schema_version SET version = 57 WHERE version = 56;
