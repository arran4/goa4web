package migrations

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0096RetriesAfterStep4Failure reproduces the state left behind by
// MariaDB when migration 0096 has already completed its earlier DDL/DML but the
// generated-column ALTER fails. MySQL/MariaDB DDL is not transactionally rolled
// back as a unit, so a patched release must be able to resume from this state.
func TestMigration0096RetriesAfterStep4Failure(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL/MariaDB migration execution tests")
	}

	db := openTemporaryMySQLDatabase(t, dsn)
	if _, err := db.Exec(`CREATE TABLE external_links (
		id INT NOT NULL AUTO_INCREMENT,
		url TEXT NOT NULL,
		clicks INT NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		updated_by INT DEFAULT NULL,
		card_title TINYTEXT,
		card_description TEXT,
		card_image TINYTEXT,
		card_image_cache TINYTEXT,
		favicon_cache TINYTEXT,
		card_duration TINYTEXT,
		card_upload_date TINYTEXT,
		card_author TINYTEXT,
		PRIMARY KEY (id)
	)`); err != nil {
		t.Fatalf("create partially migrated external_links: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_version (version INT NOT NULL)`); err != nil {
		t.Fatalf("create schema_version: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (95)`); err != nil {
		t.Fatalf("seed schema_version: %v", err)
	}
	// Step 2 from the failed run has already canonicalized this URL, and step 3
	// has already consolidated any duplicates.
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/article?id=1', 8)`); err != nil {
		t.Fatalf("seed partially migrated link: %v", err)
	}

	contents, err := FS.ReadFile("0096_mysql.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	upSection := string(contents)
	if idx := strings.Index(upSection, "-- +goose Down"); idx != -1 {
		upSection = upSection[:idx]
	}
	for _, statement := range strings.Split(upSection, ";") {
		statement = strings.TrimSpace(strings.ReplaceAll(statement, "-- +goose Up", ""))
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("execute retry Up statement %q: %v", statement, err)
		}
	}

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != 96 {
		t.Fatalf("schema version = %d, want 96", version)
	}

	var hashColumnCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'external_links' AND column_name = 'url_hash'`).Scan(&hashColumnCount); err != nil {
		t.Fatalf("read url_hash column: %v", err)
	}
	if hashColumnCount != 1 {
		t.Fatalf("url_hash column count = %d, want 1", hashColumnCount)
	}

	var hashIndexCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'external_links' AND index_name = 'external_links_url_hash_idx'`).Scan(&hashIndexCount); err != nil {
		t.Fatalf("read url hash index: %v", err)
	}
	if hashIndexCount != 1 {
		t.Fatalf("external_links_url_hash_idx count = %d, want 1", hashIndexCount)
	}
}
