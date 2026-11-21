package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMySQLMigration0107PreservesIndexedTextAndColumnCollation(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to exercise the MySQL indexed text migration")
	}
	conn := openTemporaryMySQLDatabase(t, dsn)
	ctx := t.Context()
	_, err := conn.ExecContext(ctx, `
		CREATE TABLE schema_version (version INT NOT NULL);
		INSERT INTO schema_version VALUES (106);
		CREATE TABLE roles (name VARCHAR(255) NOT NULL UNIQUE);
		CREATE TABLE users (username TINYTEXT DEFAULT NULL, UNIQUE KEY users_username_idx (username(255)));
		CREATE TABLE searchwordlist (word TINYTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL);
		CREATE TABLE content_public_labels (label TINYTEXT NOT NULL, UNIQUE KEY public_label (label(255)));
		CREATE TABLE content_private_labels (label TINYTEXT NOT NULL, UNIQUE KEY private_label (label(255)));
		CREATE TABLE content_label_status (
			item TINYTEXT NOT NULL, label TINYTEXT NOT NULL, UNIQUE KEY label_status (item(255), label(255))
		);
		INSERT INTO searchwordlist (word) VALUES ('é🪴'), ('é🪴');
		INSERT INTO users (username) VALUES (CONCAT(REPEAT('x', 254), 'a')), (CONCAT(REPEAT('x', 254), 'b'));
	`)
	require.NoError(t, err)
	migration, err := FS.ReadFile("0107_mysql.sql")
	require.NoError(t, err)
	up, down, ok := strings.Cut(string(migration), "-- +goose Down")
	require.True(t, ok)

	for range 2 {
		_, err := conn.ExecContext(ctx, up)
		require.NoError(t, err, "the migration must be safe to retry after a DDL failure")
	}
	var version int
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
	require.Equal(t, 107, version)

	var charset, collation, dataType string
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT CHARACTER_SET_NAME, COLLATION_NAME, DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'searchwordlist' AND COLUMN_NAME = 'word'`).
		Scan(&charset, &collation, &dataType))
	require.Equal(t, "utf8mb4", charset)
	require.Equal(t, "utf8mb4_bin", collation)
	require.Equal(t, "varchar", dataType)
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM searchwordlist WHERE word = 'é🪴'").Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE CHAR_LENGTH(username) = 255").Scan(&count))
	require.Equal(t, 2, count)
	_, err = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES (CONCAT(REPEAT('x', 254), 'b'))")
	var duplicate *mysql.MySQLError
	require.ErrorAs(t, err, &duplicate)
	require.EqualValues(t, 1062, duplicate.Number)

	_, err = conn.ExecContext(ctx, down)
	require.NoError(t, err)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
	require.Equal(t, 106, version)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM searchwordlist WHERE word = 'é🪴'").Scan(&count))
	require.Equal(t, 2, count)
}
