package migrations

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestMigrationFileNaming(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Regex to match NNNN.<driver>.sql
	// NNNN is 4 digits
	validNameStrict := regexp.MustCompile(`^\d{4}\.(mysql)\.sql$`)

	// Regex for files with descriptions (temporarily disallowed)
	// Matches NNNN_description.sql or NNNN_description.mysql.sql
	validNameDesc := regexp.MustCompile(`^\d{4}_[a-zA-Z0-9_.]+\.sql$`)

	// Map to track versions found for mysql driver (including generic .sql)
	// version -> filename
	mysqlVersions := make(map[int]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "embed.go" {
			continue
		}

		// Check for description usage (disallowed for now)
		if validNameDesc.MatchString(name) {
			t.Errorf("Migration file %s uses a description which is currently disabled. Use format NNNN.mysql.sql", name)
		}

		// Validate naming convention
		if !validNameStrict.MatchString(name) {
			// If it's not the strict format and not the description format (already reported), report invalid format
			if !validNameDesc.MatchString(name) {
				t.Errorf("Migration file %s does not match naming convention NNNN.mysql.sql", name)
			}
		}

		// Extract version to check for collisions
		// We assume first 4 chars are digits based on regex pass, but let's be safe
		if len(name) >= 4 {
			versionPart := name[:4]
			version, err := strconv.Atoi(versionPart)
			if err == nil {
				// Check if this file is applicable to mysql
				isMysqlApplicable := strings.HasSuffix(name, ".mysql.sql") || (strings.HasSuffix(name, ".sql") && !strings.Contains(strings.TrimSuffix(name, ".sql"), "."))

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
