-- +goose Up
CREATE TABLE IF NOT EXISTS admin_request_queue (
id INTEGER PRIMARY KEY AUTOINCREMENT,
users_idusers INT NOT NULL,
change_table TEXT NOT NULL,
change_field TEXT NOT NULL,
change_row_id int NOT NULL,
change_value TEXT,
contact_options TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
acted_at DATETIME DEFAULT NULL
);
UPDATE schema_version SET version = 43;