CREATE TABLE IF NOT EXISTS image_cache_entries (
  id varchar(255) NOT NULL,
  source_url text DEFAULT NULL,
  source_kind varchar(32) NOT NULL DEFAULT 'unknown',
  status varchar(32) NOT NULL DEFAULT 'ready',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at datetime DEFAULT NULL,
  fetched_at datetime DEFAULT NULL,
  expires_at datetime DEFAULT NULL,
  content_expires_at datetime DEFAULT NULL,
  content_type varchar(128) DEFAULT NULL,
  size_bytes bigint DEFAULT NULL,
  width int DEFAULT NULL,
  height int DEFAULT NULL,
  checksum varchar(128) DEFAULT NULL,
  thumbnail_id varchar(255) DEFAULT NULL,
  error_message text DEFAULT NULL,
  retry_count int NOT NULL DEFAULT 0,
  last_attempt_at datetime DEFAULT NULL,
  next_attempt_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY image_cache_entries_source_kind_expires_idx (source_kind, expires_at),
  KEY image_cache_entries_status_created_idx (status, created_at),
  KEY image_cache_entries_last_used_idx (last_used_at),
  KEY image_cache_entries_next_attempt_idx (status, next_attempt_at)
);

-- Record upgrade to schema version 85
UPDATE schema_version SET version = 85 WHERE version = 84;
