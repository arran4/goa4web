-- +goose Up
INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'blogs', 'entry', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'blogs', 'entry', 'allow', 'view', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'writing', 'category', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'writing', 'category', 'allow', 'view', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'writing', 'article', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'writing', 'article', 'allow', 'view', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'news', 'post', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'news', 'post', 'allow', 'view', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'faq', 'category', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'faq', 'category', 'allow', 'view', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'faq', 'question/answer', 'allow', 'see', 1;

INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, action, active)
SELECT CURRENT_TIMESTAMP, NULL, NULL, 'faq', 'question/answer', 'allow', 'view', 1;

UPDATE schema_version SET version = 57;