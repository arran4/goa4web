package migrations

import (
	"regexp"
	"strings"
	"testing"
)

func TestMigrationFileNaming(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Regex to match NNNN.<driver>.sql OR NNNN_description.sql
	// NNNN is 4 digits
	validNameStrict := regexp.MustCompile(`^\d{4}\.(mysql)\.sql$`)
	// Allow dots in description part (e.g. for .mysql.sql suffix or just dots in description)
	// This matches NNNN_something.sql
	validNameDesc := regexp.MustCompile(`^\d{4}_[a-zA-Z0-9_.]+\.sql$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "embed.go" {
			continue
		}

		isValid := false

		if validNameStrict.MatchString(name) {
			isValid = true
			if !strings.Contains(name, ".mysql.sql") {
				t.Errorf("Migration file %s matched strict regex but missing mysql suffix (should be impossible)", name)
			}
		} else if validNameDesc.MatchString(name) {
			isValid = true
			// Description files are generic sql, so no driver suffix required
		}

		if !isValid {
			t.Errorf("Migration file %s does not match naming convention NNNN.mysql.sql or NNNN_description.sql", name)
		}
	}
}
