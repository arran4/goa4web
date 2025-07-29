-- Admin email association request queue
CREATE TABLE admin_request_queue (
  id INT NOT NULL AUTO_INCREMENT,
  users_idusers INT NOT NULL,
  change_table varchar(255) NOT NULL,
  change_field varchar(255) NOT NULL,
  change_row_id int NOT NULL,
  change_value text,
  contact_options text,
  status varchar(20) NOT NULL DEFAULT 'pending',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  acted_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY admin_request_queue_user_idx (users_idusers)
);

-- Update schema version
UPDATE schema_version SET version = 43 WHERE version = 42;
