package dbstart

import (
	"bytes"
	"context"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/handlers"
)

func TestEnsureSchemaLogsVersion(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Mock expectations for SchemaVersion
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"version_id", "is_applied"}).AddRow(handlers.ExpectedSchemaVersion, true))

	// Capture log output
	var buf bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(originalWriter)
	}()

	if err := EnsureSchema(context.Background(), conn); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	output := buf.String()
	expected := "Current schema version: "
	if !strings.Contains(output, expected) {
		t.Errorf("Expected log to contain %q, but got %q", expected, output)
	}
}
