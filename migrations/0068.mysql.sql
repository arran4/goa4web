CREATE TABLE content_read_markers (
    id INT NOT NULL AUTO_INCREMENT,
    item VARCHAR(64) NOT NULL,
    item_id INT NOT NULL,
    user_id INT NOT NULL,
    last_comment_id INT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY content_read_markers_uq (item, item_id, user_id)
);

-- Update schema version
UPDATE schema_version SET version = 68 WHERE version = 67;
