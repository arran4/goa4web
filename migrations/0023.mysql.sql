ALTER TABLE writing
    CHANGE COLUMN writingCategory_idwritingCategory writing_category_id INT NOT NULL;
ALTER TABLE writing_category
    CHANGE COLUMN writingCategory_idwritingCategory writing_category_id INT NOT NULL;

-- Set default timestamp for blog posts
ALTER TABLE blogs
    MODIFY written DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Record upgrade to schema version 23
UPDATE schema_version SET version = 23 WHERE version = 22;
