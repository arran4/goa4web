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
	validNameDesc := regexp.MustCompile(`^\d{4}_[a-zA-Z0-9_]+\.sql$`)

	mysqlVersions := make(map[int]string)
	sqliteVersions := make(map[int]string)
	var maxMySQL, maxSQLite int

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
			t.Errorf("Migration file %s does not match naming convention NNNN_mysql.sql or NNNN_sqlite.sql", name)
		}

		// Extract version to track independently
		if len(name) >= 4 {
			versionPart := name[:4]
			version, err := strconv.Atoi(versionPart)
			if err == nil {
				if strings.HasSuffix(name, "_mysql.sql") {
					if existingFile, exists := mysqlVersions[version]; exists {
						t.Errorf("Duplicate MySQL migration version %d found: %s and %s", version, existingFile, name)
					}
					mysqlVersions[version] = name
					if version > maxMySQL {
						maxMySQL = version
					}
				} else if strings.HasSuffix(name, "_sqlite.sql") {
					if existingFile, exists := sqliteVersions[version]; exists {
						t.Errorf("Duplicate SQLite migration version %d found: %s and %s", version, existingFile, name)
					}
					sqliteVersions[version] = name
					if version > maxSQLite {
						maxSQLite = version
					}
				}
			}
		}
	}

	// Verify parity between MySQL and SQLite migrations
	if maxMySQL != maxSQLite {
		t.Errorf("Migration version mismatch: max MySQL version is %d, max SQLite version is %d", maxMySQL, maxSQLite)
	}

	for v := 1; v <= maxMySQL; v++ {
		if _, ok := mysqlVersions[v]; !ok {
			t.Errorf("Missing MySQL migration for version %04d", v)
		}
		if _, ok := sqliteVersions[v]; !ok {
			t.Errorf("Missing SQLite migration for version %04d", v)
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

	for _, schemaFile := range []string{"schema.mysql.sql", "schema.sqlite.sql"} {
		schemaPath := filepath.Join("..", "database", schemaFile)
		content, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Fatalf("Failed to read schema file at %s: %v", schemaPath, err)
		}

		schemaStr := string(content)
		expected := fmt.Sprintf("(%d, 1)", maxVersion)
		if !strings.Contains(schemaStr, expected) || !strings.Contains(schemaStr, "goose_db_version") {
			t.Errorf("Schema file %s does not contain expected version update for %d", schemaPath, maxVersion)
		}
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

func TestMigration0096ExecutesSuccessfully(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL migration execution tests")
	}
	db := openTemporaryMySQLDatabase(t, dsn)
	pre0096Table := `CREATE TABLE external_links (
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
		PRIMARY KEY (id),
		UNIQUE KEY external_links_url_idx (url(255))
	)`
	if _, err := db.Exec(pre0096Table); err != nil {
		t.Fatalf("create pre-0096 external_links: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_version (version INT NOT NULL)`); err != nil {
		t.Fatalf("create schema_version: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (95)`); err != nil {
		t.Fatalf("seed schema_version: %v", err)
	}

	// 1. Seed pre-existing tracking variants that will collide under the old prefix index
	// Row 1 (will be keep_id) has no title/description initially
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks, card_title, card_description) VALUES ('https://example.com/article?id=1&utm_source=old', 5, NULL, NULL)`); err != nil {
		t.Fatalf("seed link 1: %v", err)
	}
	// Row 2 has title and description populated
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks, card_title, card_description) VALUES ('https://example.com/article?id=1&utm_medium=twitter', 3, 'Article 1 Title', 'Article 1 Description')`); err != nil {
		t.Fatalf("seed link 2: %v", err)
	}

	// 2. Seed query-string position cases
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/item1?utm_source=x&id=1&category=books', 1)`); err != nil {
		t.Fatalf("seed position 1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/item2?id=1&utm_source=x&category=books', 1)`); err != nil {
		t.Fatalf("seed position 2: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/item3?id=1&category=books&utm_source=x', 1)`); err != nil {
		t.Fatalf("seed position 3: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/item4?utm_source=x&utm_medium=y&id=1&category=books', 1)`); err != nil {
		t.Fatalf("seed position 4: %v", err)
	}

	// 3. Seed all tracking key families
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/track?fbclid=1&gclid=2&gbraid=3&wbraid=4&mc_cid=5&mc_eid=6&igshid=7&msclkid=8&twclid=9&yclid=10&click_id=11&clickid=12&_hsenc=13&_hsmi=14&mkt_tok=15&id=99', 1)`); err != nil {
		t.Fatalf("seed all tracking keys: %v", err)
	}

	// 4. Seed signed URLs (must remain unchanged)
	signedS3 := "https://bucket.s3.amazonaws.com/file?X-Amz-Signature=abc&utm_source=x"
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES (?, 1)`, signedS3); err != nil {
		t.Fatalf("seed signed S3: %v", err)
	}
	signedCloudFront := "https://d111111abcdef8.cloudfront.net/video.mp4?Signature=xyz&utm_medium=y"
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES (?, 1)`, signedCloudFront); err != nil {
		t.Fatalf("seed signed CloudFront: %v", err)
	}

	// 5. Seed functional and unknown parameters
	bloombergURL := "https://www.bloomberg.com/news/sample?accessToken=token123&resource=finance&X-Amz-Security-Token=sec456&custom_param=val&utm_source=email"
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES (?, 1)`, bloombergURL); err != nil {
		t.Fatalf("seed bloomberg URL: %v", err)
	}

	// 6. Seed empty query components and percent-encoded keys
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/empty1?a=1&&utm_source=x&b=2', 1)`); err != nil {
		t.Fatalf("seed empty components 1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/empty2?&a=1&utm_source=x', 1)`); err != nil {
		t.Fatalf("seed empty components 2: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/empty3?a=1&utm_source=x&', 1)`); err != nil {
		t.Fatalf("seed empty components 3: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/empty4?a=1&&utm_source=x&&b=2', 1)`); err != nil {
		t.Fatalf("seed empty components 4: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/encoded?%75tm_source=x&id=1', 1)`); err != nil {
		t.Fatalf("seed encoded key: %v", err)
	}

	// 7. Seed uppercase scheme, bare question marks, non-http schemes, and complex URLs
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('HTTPS://example.com/upper?id=1&utm_source=x', 1)`); err != nil {
		t.Fatalf("seed uppercase scheme: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/bare1?&utm_source=x', 1)`); err != nil {
		t.Fatalf("seed bare question mark 1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/bare2?utm_source=x&', 1)`); err != nil {
		t.Fatalf("seed bare question mark 2: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/bare3?&utm_source=x&', 1)`); err != nil {
		t.Fatalf("seed bare question mark 3: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('mailto:user@example.com?utm_source=x', 1)`); err != nil {
		t.Fatalf("seed mailto: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('ftp://example.com/file?utm_source=x', 1)`); err != nil {
		t.Fatalf("seed ftp: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('//example.com/file?utm_source=x', 1)`); err != nil {
		t.Fatalf("seed scheme-relative: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42&utm_source=x#section-1', 1)`); err != nil {
		t.Fatalf("seed complex url: %v", err)
	}

	// 8. Seed UTM prefix cases with non-alphanumeric characters and non-UTM keys
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/utm1?utm_campaign-name=x&id=1', 1)`); err != nil {
		t.Fatalf("seed utm hyphen: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/utm2?utm_custom.value=x&id=1', 1)`); err != nil {
		t.Fatalf("seed utm dot: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/utm3?UTM_custom-value=x&id=1', 1)`); err != nil {
		t.Fatalf("seed utm upper: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/utm4?utm.foo=x&id=1', 1)`); err != nil {
		t.Fatalf("seed non-utm dot: %v", err)
	}

	// 9. Seed exact-key prefix regression cases (must NOT be corrupted)
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/prefix1?id=1&gclid_extra=keep&utm_source=x', 1)`); err != nil {
		t.Fatalf("seed gclid_extra: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/prefix2?id=1&fbclid_extra=val&utm_medium=mail', 1)`); err != nil {
		t.Fatalf("seed fbclid_extra: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES ('https://example.com/prefix3?id=1&clickidfoo=123&gclid=real', 1)`); err != nil {
		t.Fatalf("seed clickidfoo: %v", err)
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

	// Verify schema_version is 96
	var currentVersion int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&currentVersion); err != nil || currentVersion != 96 {
		t.Fatalf("expected schema version 96 after Up migration, got %d, err %v", currentVersion, err)
	}

	// Verify consolidation: click sum and metadata preservation
	var count int
	var clicks int
	var title, description sql.NullString
	err = db.QueryRow(`SELECT COUNT(*), SUM(clicks), MAX(card_title), MAX(card_description) FROM external_links WHERE url = 'https://example.com/article?id=1' GROUP BY url`).Scan(&count, &clicks, &title, &description)
	if err != nil {
		t.Fatalf("query consolidated row: %v", err)
	}
	if count != 1 || clicks != 8 {
		t.Fatalf("expected 1 row with 8 clicks, got count=%d, clicks=%d", count, clicks)
	}
	if !title.Valid || title.String != "Article 1 Title" {
		t.Fatalf("expected preserved title 'Article 1 Title', got %v", title)
	}
	if !description.Valid || description.String != "Article 1 Description" {
		t.Fatalf("expected preserved description 'Article 1 Description', got %v", description)
	}

	// Verify query positions retained exact 'id=1&category=books' without '?category=books' corruption
	for i := 1; i <= 4; i++ {
		expectedURL := fmt.Sprintf("https://example.com/item%d?id=1&category=books", i)
		var matched int
		if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = ?`, expectedURL).Scan(&matched); err != nil || matched != 1 {
			t.Fatalf("expected cleaned URL %q to exist in database, matched=%d, err=%v", expectedURL, matched, err)
		}
	}

	// Verify all tracking keys stripped and 'id=99' retained
	var trackMatched int
	if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = 'https://example.com/track?id=99'`).Scan(&trackMatched); err != nil || trackMatched != 1 {
		t.Fatalf("expected 'https://example.com/track?id=99', matched=%d, err=%v", trackMatched, err)
	}

	// Verify signed URLs remained byte-for-byte untouched
	var s3Matched, cfMatched int
	if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = ?`, signedS3).Scan(&s3Matched); err != nil || s3Matched != 1 {
		t.Fatalf("expected signed S3 URL to remain untouched, matched=%d, err=%v", s3Matched, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = ?`, signedCloudFront).Scan(&cfMatched); err != nil || cfMatched != 1 {
		t.Fatalf("expected signed CloudFront URL to remain untouched, matched=%d, err=%v", cfMatched, err)
	}

	// Verify functional parameters preserved
	expectedBloomberg := "https://www.bloomberg.com/news/sample?accessToken=token123&resource=finance&X-Amz-Security-Token=sec456&custom_param=val"
	var bloombergMatched int
	if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = ?`, expectedBloomberg).Scan(&bloombergMatched); err != nil || bloombergMatched != 1 {
		t.Fatalf("expected Bloomberg URL with functional params %q, matched=%d, err=%v", expectedBloomberg, bloombergMatched, err)
	}

	// Verify empty query components, percent-encoded keys, uppercase scheme, bare question marks, non-http schemes, and complex URLs
	expectedCases := map[string]string{
		"https://example.com/empty1?a=1&&b=2":                                 "empty components between params",
		"https://example.com/empty2?&a=1":                                     "leading empty component",
		"https://example.com/empty3?a=1&":                                     "trailing empty component",
		"https://example.com/empty4?a=1&&&b=2":                                "multiple empty components",
		"https://example.com/encoded?%75tm_source=x&id=1":                     "percent-encoded key untreated",
		"HTTPS://example.com/upper?id=1":                                      "uppercase scheme preserved",
		"https://example.com/bare1?":                                          "bare question mark leading",
		"https://example.com/bare2?":                                          "bare question mark trailing",
		"https://example.com/bare3?&":                                         "bare question mark with ampersand",
		"mailto:user@example.com?utm_source=x":                                "mailto untouched",
		"ftp://example.com/file?utm_source=x":                                 "ftp untouched",
		"//example.com/file?utm_source=x":                                     "scheme-relative untouched",
		"https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42#section-1": "complex url preserved",
		"https://example.com/utm1?id=1":                                       "utm hyphen stripped",
		"https://example.com/utm2?id=1":                                       "utm dot stripped",
		"https://example.com/utm3?id=1":                                       "utm upper hyphen stripped",
		"https://example.com/utm4?utm.foo=x&id=1":                             "non-utm dot preserved",
		"https://example.com/prefix1?id=1&gclid_extra=keep":                   "gclid_extra preserved and utm_source stripped",
		"https://example.com/prefix2?id=1&fbclid_extra=val":                   "fbclid_extra preserved and utm_medium stripped",
		"https://example.com/prefix3?id=1&clickidfoo=123":                     "clickidfoo preserved and gclid stripped",
	}
	for expURL, desc := range expectedCases {
		var matched int
		if err := db.QueryRow(`SELECT COUNT(*) FROM external_links WHERE url = ?`, expURL).Scan(&matched); err != nil || matched != 1 {
			t.Fatalf("expected %s URL %q to exist in database, matched=%d, err=%v", desc, expURL, matched, err)
		}
	}

	// Test two URLs > 255 chars with identical first 255 chars
	prefix := "https://example.com/very/long/path/" + strings.Repeat("a", 230) + "?common=true&suffix="
	url1 := prefix + "one"
	url2 := prefix + "two"

	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES (?, 10)`, url1); err != nil {
		t.Fatalf("insert long url 1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO external_links (url, clicks) VALUES (?, 20)`, url2); err != nil {
		t.Fatalf("insert long url 2: %v", err)
	}

	var id1, id2, clicks1, clicks2 int
	if err := db.QueryRow(`SELECT id, clicks FROM external_links WHERE url_hash = UNHEX(SHA2(?, 256))`, url1).Scan(&id1, &clicks1); err != nil {
		t.Fatalf("query long url 1: %v", err)
	}
	if err := db.QueryRow(`SELECT id, clicks FROM external_links WHERE url_hash = UNHEX(SHA2(?, 256))`, url2).Scan(&id2, &clicks2); err != nil {
		t.Fatalf("query long url 2: %v", err)
	}
	if id1 == id2 || clicks1 != 10 || clicks2 != 20 {
		t.Fatalf("long urls collided: id1=%d, id2=%d, clicks1=%d, clicks2=%d", id1, id2, clicks1, clicks2)
	}

	// Test guarded Down rollback with >255-byte URLs
	downSection := rawSQL[strings.Index(rawSQL, "-- +goose Down"):]
	guardStmt := ""
	for _, stmt := range strings.Split(downSection, ";") {
		stmt = strings.TrimSpace(strings.ReplaceAll(stmt, "-- +goose Down", ""))
		if strings.Contains(stmt, "@guard") {
			guardStmt = stmt
			break
		}
	}
	if guardStmt == "" {
		t.Fatal("missing guard statement in Down section")
	}

	// Guard must fail because long URLs exist
	if _, err := db.Exec(guardStmt); err == nil {
		t.Fatal("expected Down guard statement to fail when URLs > 255 bytes exist")
	}

	// Verify long URLs were not truncated or deleted
	var fullURL string
	if err := db.QueryRow(`SELECT url FROM external_links WHERE id = ?`, id1).Scan(&fullURL); err != nil || fullURL != url1 {
		t.Fatalf("long URL was damaged after guard failure: %v, got %s", err, fullURL)
	}

	// Delete long URLs and verify rollback succeeds
	if _, err := db.Exec(`DELETE FROM external_links WHERE LENGTH(url) > 255`); err != nil {
		t.Fatalf("delete long urls: %v", err)
	}

	for _, statement := range strings.Split(downSection, ";") {
		statement = strings.TrimSpace(strings.ReplaceAll(statement, "-- +goose Down", ""))
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("execute Down statement %q: %v", statement, err)
		}
	}

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version); err != nil || version != 95 {
		t.Fatalf("expected rolled back schema version 95, got %d, err %v", version, err)
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
	if _, err := adminDB.Exec("CREATE DATABASE `" + databaseName + "` DEFAULT CHARACTER SET latin1"); err != nil {
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
