package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestMigrationFileNaming(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Regex to match NNNN_<driver>.sql
	// NNNN is 4 digits
	validNameStrict := regexp.MustCompile(`^\d{4}_(mysql)\.sql$`)

	// Regex for files with descriptions (temporarily disallowed)
	// Matches NNNN_description.sql or NNNN_description.mysql.sql
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
			t.Errorf("Migration file %s uses a description which is currently disabled. Use format NNNN_mysql.sql", name)
		}

		// Validate naming convention
		if !validNameStrict.MatchString(name) {
			// If it's not the strict format and not the description format (already reported), report invalid format
			if !validNameDesc.MatchString(name) {
				t.Errorf("Migration file %s does not match naming convention NNNN_mysql.sql", name)
			}
		}

		// Extract version to check for collisions
		// We assume first 4 chars are digits based on regex pass, but let's be safe
		if len(name) >= 4 {
			versionPart := name[:4]
			version, err := strconv.Atoi(versionPart)
			if err == nil {
				// Check if this file is applicable to mysql
				isMysqlApplicable := strings.HasSuffix(name, "_mysql.sql") || (strings.HasSuffix(name, ".sql") && !strings.Contains(strings.TrimSuffix(name, ".sql"), "_"))

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
	// We expect the line: INSERT INTO `goose_db_version` (`version_id`, `is_applied`) VALUES (89, 1);
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
	wantWhere := fmt.Sprintf("WHERE section = '%s' AND item = '%s'", consts.PermissionSectionPrivateForum, consts.PermissionItemThread)
	if !strings.Contains(sql, wantWhere) {
		t.Error("migration is not restricted to legacy private thread grants")
	}
	if !strings.Contains(sql, "JOIN forumthread thread_row") {
		t.Error("migration does not backfill grants for existing private threads")
	}
	if !strings.Contains(sql, "topic_grant.action IN ('see', 'view', 'post', 'reply')") {
		t.Error("migration does not restrict backfilled grants to supported private thread actions")
	}
	if !strings.Contains(sql, "AND NOT EXISTS") {
		t.Error("migration does not guard against duplicate private thread grants")
	}
}
