-- Comments for admin request queue entries
CREATE TABLE admin_request_comments (
  id INT NOT NULL AUTO_INCREMENT,
  request_id INT NOT NULL,
  comment TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY admin_request_comments_request_idx (request_id)
);

-- Update schema version
UPDATE schema_version SET version = 44 WHERE version = 43;
