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

	// Regex to match NNNN.<driver>.sql
	// Assuming NNNN is 4 digits
	validName := regexp.MustCompile(`^\d{4}\.(mysql)\.sql$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "embed.go" {
			continue
		}
		if !validName.MatchString(name) {
			t.Errorf("Migration file %s does not match naming convention NNNN.driver.sql (e.g. 0001.mysql.sql)", name)
		}

		// Also verify driver is mysql
		if !strings.Contains(name, ".mysql.sql") {
			t.Errorf("Migration file %s must use mysql driver suffix", name)
		}
	}
}
