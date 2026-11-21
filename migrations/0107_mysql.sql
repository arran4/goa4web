-- +goose Up
-- +goose StatementBegin
-- A full TINYTEXT prefix is invalid under MySQL utf8mb4.
-- Keep each column's existing character set and collation, including when
-- they differ from the table defaults. Existing VARCHAR columns need no DDL.
SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE roles MODIFY COLUMN name VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' NOT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'roles' AND COLUMN_NAME = 'name'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE users MODIFY COLUMN username VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' DEFAULT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'username'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE searchwordlist MODIFY COLUMN word VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' DEFAULT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'searchwordlist' AND COLUMN_NAME = 'word'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE content_public_labels MODIFY COLUMN label VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' NOT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'content_public_labels' AND COLUMN_NAME = 'label'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE content_private_labels MODIFY COLUMN label VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' NOT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'content_private_labels' AND COLUMN_NAME = 'label'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE content_label_status MODIFY COLUMN item VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' NOT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'content_label_status' AND COLUMN_NAME = 'item'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

SET @goa4web_indexed_text_ddl = (
    SELECT IF(DATA_TYPE = 'tinytext',
        CONCAT('ALTER TABLE content_label_status MODIFY COLUMN label VARCHAR(255) CHARACTER SET ',
            CHARACTER_SET_NAME, ' COLLATE ', COLLATION_NAME, ' NOT NULL'),
        'DO 0')
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'content_label_status' AND COLUMN_NAME = 'label'
);
PREPARE goa4web_indexed_text_stmt FROM @goa4web_indexed_text_ddl;
EXECUTE goa4web_indexed_text_stmt;
DEALLOCATE PREPARE goa4web_indexed_text_stmt;

UPDATE schema_version SET version = 107;
-- +goose StatementEnd

-- +goose Down
-- Earlier application versions support these columns. Retain the widened
-- representation because converting back could lose values and fail on utf8mb4.
UPDATE schema_version SET version = 106;
