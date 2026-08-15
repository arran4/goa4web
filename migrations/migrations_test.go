package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"
	"github.com/go-sql-driver/mysql"
)

func TestMigrationFileNaming(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Regex to match NNNN_<driver>.sql
	// NNNN is 4 digits
	validNameStrict := regexp.MustCompile(`^\d{4}_(?:mysql|sqlite)\.sql$`)

	// Regex for files with descriptions (temporarily disallowed)
	// Matches NNNN_description.sql or NNNN_description.mysql.sql|sqlite.sql
	validNameDesc := regexp.MustCompile(`^\d{4}_[a-zA-Z0-9_]+\.sql$`)

	// Map to track versions found for mysql driver (including generic .sql)
	// version -> filename
	mysqlVersions := make(map[int]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "embed.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Check for description usage (disallowed for now)
		if validNameDesc.MatchString(name) && !validNameStrict.MatchString(name) {
			t.Errorf("Migration file %s uses a description which is currently disabled. Use format NNNN_mysql.sql or NNNN_sqlite.sql", name)
		}

		// Validate naming convention
		if !validNameStrict.MatchString(name) {
			// If it's not the strict format and not the description format (already reported), report invalid format
			if !validNameDesc.MatchString(name) {
				t.Errorf("Migration file %s does not match naming convention NNNN_mysql.sql or NNNN_sqlite.sql|sqlite.sql", name)
			}
		}

		// Extract version to check for collisions
		// We assume first 4 chars are digits based on regex pass, but let's be safe
		if len(name) >= 4 {
			versionPart := name[:4]
			version, err := strconv.Atoi(versionPart)
			if err == nil {
				// Check if this file is applicable to mysql
				isMysqlApplicable := strings.HasSuffix(name, "_mysql.sql|sqlite.sql") || (strings.HasSuffix(name, ".sql") && !strings.Contains(strings.TrimSuffix(name, ".sql"), "_"))

				if isMysqlApplicable {
					if existingFile, exists := mysqlVersions[version]; exists {
						t.Errorf("Duplicate migration version %d found: %s and %s are mutually exclusive for mysql", version, existingFile, name)
					} else {
						mysqlVersions[version] = name
					}
				}
			}
		}
	}
}

func TestSchemaVersionUpdated(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	var maxVersion int
	for _, entry := range entries {
		name := entry.Name()
		if len(name) >= 4 {
			if version, err := strconv.Atoi(name[:4]); err == nil {
				if version > maxVersion {
					maxVersion = version
				}
			}
		}
	}

	if maxVersion == 0 {
		t.Skip("No migrations found")
	}

	schemaPath := filepath.Join("..", "database", "schema.mysql.sql")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file at %s: %v", schemaPath, err)
	}

	schemaStr := string(content)
	// We expect the line: INSERT INTO `goose_db_version` (`version_id`, `is_applied`) VALUES (90, 1);
	// Allow for some whitespace variation
	expected := fmt.Sprintf("INSERT INTO `goose_db_version` (`version_id`, `is_applied`) VALUES (%d, 1)", maxVersion)
	if !strings.Contains(schemaStr, expected) {
		t.Errorf("Schema file %s does not contain expected version update:\nExpected substring: %s\nEnsure you have updated the goose_db_version insert in database/schema.mysql.sql", schemaPath, expected)
	}
}

func TestPrivateForumThreadGrantMigration(t *testing.T) {
	content, err := FS.ReadFile("0094_mysql.sql")
	if err != nil {
		t.Fatalf("read private forum thread grant migration: %v", err)
	}

	sql := string(content)
	wantSet := fmt.Sprintf("SET section = '%s'", consts.PermissionSectionPrivateForumThread)
	if !strings.Contains(sql, wantSet) {
		t.Error("migration does not move grants to privateforum_thread")
	}
	wantSection := fmt.Sprintf("WHERE section = '%s'", consts.PermissionSectionPrivateForum)
	wantItem := fmt.Sprintf("AND item = '%s'", consts.PermissionItemThread)
	if !strings.Contains(sql, wantSection) || !strings.Contains(sql, wantItem) {
		t.Error("migration is not restricted to legacy private thread grants")
	}
	if !strings.Contains(sql, "JOIN forumthread thread_row") {
		t.Error("migration does not backfill grants for existing private threads")
	}
	if !strings.Contains(sql, "topic_grant.action IN ('view', 'reply')") {
		t.Error("migration does not restrict backfilled grants to supported private thread actions")
	}
	if !strings.Contains(sql, "rule_type = 'allow'") || !strings.Contains(sql, "active = 1") || !strings.Contains(sql, "item_id IS NOT NULL") {
		t.Error("migration does not restrict normalized legacy grants to active, valid allow rows")
	}
	if !strings.Contains(sql, "AND NOT EXISTS") {
		t.Error("migration does not guard against duplicate private thread grants")
	}
	if !strings.Contains(sql, "thread_grant.rule_type = 'allow'") {
		t.Error("migration treats non-allow grants as equivalent to the allow grants it backfills")
	}
	if !strings.Contains(sql, "UPDATE schema_version SET version = 94") {
		t.Error("migration does not update the legacy schema version")
	}
}

func TestMigration0095ExecutesSuccessfully(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL migration execution tests")
	}
	db := openTemporaryMySQLDatabase(t, dsn)
	if _, err := db.Exec(`CREATE TABLE forumthread (idforumthread INT NOT NULL AUTO_INCREMENT PRIMARY KEY)`); err != nil {
		t.Fatalf("create pre-0095 forumthread: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_version (version INT NOT NULL)`); err != nil {
		t.Fatalf("create schema_version: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (94)`); err != nil {
		t.Fatalf("seed schema_version: %v", err)
	}
	contents, err := FS.ReadFile("0095_mysql.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	for _, statement := range strings.Split(string(contents), ";") {
		statement = strings.TrimSpace(strings.ReplaceAll(statement, "-- +goose Up", ""))
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("execute migration statement %q: %v", statement, err)
		}
	}
	rows, err := db.Query(`SELECT column_name FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'forumthread' AND index_name = 'forumthread_reply_to_thread_id' ORDER BY seq_in_index`)
	if err != nil {
		t.Fatalf("show fork index: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			t.Fatalf("scan fork index: %v", err)
		}
		columns = append(columns, columnName)
	}
	if got := strings.Join(columns, ","); got != "reply_to_thread_id,reply_to_comment_id" {
		t.Fatalf("fork index columns = %q", got)
	}
}

func openTemporaryMySQLDatabase(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse MySQL DSN: %v", err)
	}
	databaseName := fmt.Sprintf("goa4web_migration_%d", time.Now().UnixNano())
	adminConfig := *cfg
	adminConfig.DBName = ""
	adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		t.Fatalf("open MySQL admin connection: %v", err)
	}
	if _, err := adminDB.Exec("CREATE DATABASE `" + databaseName + "`"); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create temporary database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.Exec("DROP DATABASE `" + databaseName + "`")
		_ = adminDB.Close()
	})
	testConfig := *cfg
	testConfig.DBName = databaseName
	db, err := sql.Open("mysql", testConfig.FormatDSN())
	if err != nil {
		t.Fatalf("open temporary database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
