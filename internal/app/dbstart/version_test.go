package dbstart

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/migrations"
)

func TestExpectedSchemaVersionMatchesMigrations(t *testing.T) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(name, ".sql"))
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	if max != handlers.ExpectedSchemaVersion {
		t.Fatalf("schema version constant %d does not match latest migration %d", handlers.ExpectedSchemaVersion, max)
	}
}
