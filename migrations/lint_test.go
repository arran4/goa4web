package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers"
)

func TestMigrationsLint(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}

	maxByDriver := map[string]int{"mysql": 0, "sqlite": 0}
	type migrationEntry struct {
		driver  string
		version int
		name    string
	}
	var checked []migrationEntry

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") || len(name) < 5 {
			continue
		}
		ver, err := strconv.Atoi(name[:4])
		if err != nil {
			continue
		}
		var driver string
		switch {
		case strings.HasSuffix(name, "_mysql.sql"):
			driver = "mysql"
		case strings.HasSuffix(name, "_sqlite.sql"):
			driver = "sqlite"
		default:
			continue
		}
		if ver > maxByDriver[driver] {
			maxByDriver[driver] = ver
		}
		if ver >= 99 {
			checked = append(checked, migrationEntry{driver: driver, version: ver, name: name})
		}
	}

	for _, driver := range []string{"mysql", "sqlite"} {
		if maxByDriver[driver] != handlers.ExpectedSchemaVersion {
			t.Errorf("%s max migration version = %d, ExpectedSchemaVersion = %d", driver, maxByDriver[driver], handlers.ExpectedSchemaVersion)
		}
	}

	gooseMarker := regexp.MustCompile(fmt.Sprintf(`\(\s*%d\s*,\s*1\s*\)`, handlers.ExpectedSchemaVersion))
	legacyMarker := regexp.MustCompile(fmt.Sprintf(`(?i)INSERT\s+INTO\s+`+"`?"+`schema_version`+"`?"+`\s*\(\s*`+"`?"+`version`+"`?"+`\s*\)\s*VALUES\s*\(\s*%d\s*\)`, handlers.ExpectedSchemaVersion))
	for _, schemaFile := range []string{"schema.mysql.sql", "schema.sqlite.sql"} {
		path := filepath.Join("..", "database", schemaFile)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(raw)
		if !strings.Contains(content, "goose_db_version") || !gooseMarker.MatchString(content) {
			t.Errorf("%s missing goose_db_version marker (%d, 1)", schemaFile, handlers.ExpectedSchemaVersion)
		}
		if !strings.Contains(content, "schema_version") || !legacyMarker.MatchString(content) {
			t.Errorf("%s missing schema_version marker (%d)", schemaFile, handlers.ExpectedSchemaVersion)
		}
	}

	for _, m := range checked {
		raw, err := FS.ReadFile(m.name)
		if err != nil {
			t.Fatalf("read %s: %v", m.name, err)
		}
		sqlText := string(raw)
		upIdx := strings.Index(sqlText, "-- +goose Up")
		downIdx := strings.Index(sqlText, "-- +goose Down")
		if upIdx < 0 || downIdx < 0 || upIdx >= downIdx {
			t.Errorf("%s must contain '-- +goose Up' before '-- +goose Down'", m.name)
			continue
		}
		upSection := sqlText[upIdx:downIdx]
		downSection := sqlText[downIdx:]

		checkGooseStatementBlocks(t, m.name, "Up", upSection, m.driver == "mysql")
		checkGooseStatementBlocks(t, m.name, "Down", downSection, m.driver == "mysql")

		if m.driver == "mysql" {
			wantUp := fmt.Sprintf("UPDATE schema_version SET version = %d", m.version)
			wantDown := fmt.Sprintf("UPDATE schema_version SET version = %d", m.version-1)
			if !strings.Contains(upSection, wantUp) {
				t.Errorf("%s Up section missing %q", m.name, wantUp)
			}
			if !strings.Contains(downSection, wantDown) {
				t.Errorf("%s Down section missing %q", m.name, wantDown)
			}
		}
	}
}

func checkGooseStatementBlocks(t *testing.T, filename, sectionName, section string, requirePreparedBlock bool) {
	t.Helper()
	inBlock := false
	for lineNum, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "-- +goose StatementBegin":
			if inBlock {
				t.Errorf("%s (%s:%d): nested StatementBegin block", filename, sectionName, lineNum+1)
			}
			inBlock = true
		case "-- +goose StatementEnd":
			if !inBlock {
				t.Errorf("%s (%s:%d): unmatched StatementEnd block", filename, sectionName, lineNum+1)
			}
			inBlock = false
		default:
			if requirePreparedBlock && !inBlock {
				upper := strings.ToUpper(trimmed)
				if strings.HasPrefix(upper, "PREPARE ") || strings.HasPrefix(upper, "EXECUTE ") || strings.HasPrefix(upper, "DEALLOCATE ") {
					t.Errorf("%s (%s:%d): prepared statement %q must be inside StatementBegin/StatementEnd", filename, sectionName, lineNum+1, trimmed)
				}
			}
		}
	}
	if inBlock {
		t.Errorf("%s (%s): unclosed StatementBegin block", filename, sectionName)
	}
}
