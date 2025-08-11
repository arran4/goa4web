CREATE TABLE content_read_markers (
    id SERIAL PRIMARY KEY,
    item TEXT NOT NULL,
    item_id INT NOT NULL,
    user_id INT NOT NULL,
    last_comment_id INT NOT NULL,
    UNIQUE (item, item_id, user_id)
);

-- Update schema version
UPDATE schema_version SET version = 68 WHERE version = 67;
