CREATE TABLE IF NOT EXISTS thread_images (
    idthread_image INT NOT NULL AUTO_INCREMENT,
    forumthread_id INT NOT NULL,
    path TINYTEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (idthread_image),
    KEY thread_images_thread_idx (forumthread_id)
);

-- Record upgrade to schema version 77
UPDATE schema_version SET version = 77 WHERE version = 76;
