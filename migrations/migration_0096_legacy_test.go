package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestMigration0096ExecutesWithoutLegacyURLIndex(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL migration execution tests")
	}

	db := openTemporaryMySQLDatabase(t, dsn)
	if _, err := db.Exec(`CREATE TABLE external_links (
		id INT NOT NULL AUTO_INCREMENT,
		url TINYTEXT NOT NULL,
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
		t.Fatalf("create legacy external_links without url index: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_version (version INT NOT NULL)`); err != nil {
		t.Fatalf("create schema_version: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (95)`); err != nil {
		t.Fatalf("seed schema_version: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/article?id=1&utm_source=legacy', 1)`); err != nil {
		t.Fatalf("seed legacy external link: %v", err)
	}

	contents, err := FS.ReadFile("0096_mysql.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	rawSQL := string(contents)
	upSection := rawSQL
	if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
		upSection = rawSQL[:idx]
	}
	for _, statement := range strings.Split(upSection, ";") {
		statement = strings.TrimSpace(strings.ReplaceAll(statement, "-- +goose Up", ""))
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("execute Up statement %q: %v", statement, err)
		}
	}

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != 96 {
		t.Fatalf("schema version = %d, want 96", version)
	}

	var dataType string
	if err := db.QueryRow(`SELECT data_type FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'external_links' AND column_name = 'url'`).Scan(&dataType); err != nil {
		t.Fatalf("read url column type: %v", err)
	}
	if dataType != "text" {
		t.Fatalf("url data type = %q, want text", dataType)
	}

	var hashIndexCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'external_links' AND index_name = 'external_links_url_hash_idx'`).Scan(&hashIndexCount); err != nil {
		t.Fatalf("read url hash index: %v", err)
	}
	if hashIndexCount != 1 {
		t.Fatalf("external_links_url_hash_idx count = %d, want 1", hashIndexCount)
	}

	var cleanedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = 'https://example.com/article?id=1'`).Scan(&cleanedCount); err != nil {
		t.Fatalf("read cleaned external link: %v", err)
	}
	if cleanedCount != 1 {
		t.Fatalf("cleaned external link count = %d, want 1", cleanedCount)
	}
}
