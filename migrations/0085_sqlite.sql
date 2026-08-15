-- +goose Up
CREATE TABLE IF NOT EXISTS image_cache_entries (
id TEXT NOT NULL,
source_url TEXT DEFAULT NULL,
source_kind TEXT NOT NULL DEFAULT 'unknown',
status TEXT NOT NULL DEFAULT 'ready',
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
last_used_at DATETIME DEFAULT NULL,
fetched_at DATETIME DEFAULT NULL,
expires_at DATETIME DEFAULT NULL,
content_expires_at DATETIME DEFAULT NULL,
content_type TEXT DEFAULT NULL,
size_bytes bigint DEFAULT NULL,
width int DEFAULT NULL,
height int DEFAULT NULL,
checksum TEXT DEFAULT NULL,
thumbnail_id TEXT DEFAULT NULL,
error_message TEXT DEFAULT NULL,
retry_count int NOT NULL DEFAULT 0,
last_attempt_at DATETIME DEFAULT NULL,
next_attempt_at DATETIME DEFAULT NULL,
PRIMARY KEY (id)
);