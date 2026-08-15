-- +goose Up
CREATE TABLE grants (
id INTEGER PRIMARY KEY AUTOINCREMENT,
created_at DATETIME NULL,
updated_at DATETIME NULL,
user_id INT NULL,
role_id INT NULL,
section TEXT NOT NULL,
item TEXT NULL,
rule_type TEXT NOT NULL,
item_id INT NULL,
item_rule TEXT NULL,
action TEXT NOT NULL,
extra TEXT NULL,
active INTEGER NOT NULL DEFAULT 1
);