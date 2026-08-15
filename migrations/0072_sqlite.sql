-- +goose Up
CREATE TABLE role_subscription_archetypes (
id INTEGER PRIMARY KEY AUTOINCREMENT,
role_id int NOT NULL,
archetype_name TEXT NOT NULL,
pattern TEXT NOT NULL,
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);