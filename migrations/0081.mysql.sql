DROP PROCEDURE IF EXISTS SafeAddColumn;
CREATE PROCEDURE SafeAddColumn(IN tableName VARCHAR(255), IN colName VARCHAR(255), IN colDef TEXT)
BEGIN
    IF NOT EXISTS (
        SELECT * FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = tableName
        AND COLUMN_NAME = colName
    ) THEN
        SET @stmt = CONCAT('ALTER TABLE ', tableName, ' ADD COLUMN ', colName, ' ', colDef);
        PREPARE stmt FROM @stmt;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END;

CALL SafeAddColumn('faq', 'deleted_at', 'DATETIME DEFAULT NULL');
CALL SafeAddColumn('faq_categories', 'deleted_at', 'DATETIME DEFAULT NULL');
CALL SafeAddColumn('forumcategory', 'deleted_at', 'DATETIME DEFAULT NULL');

DROP PROCEDURE SafeAddColumn;
