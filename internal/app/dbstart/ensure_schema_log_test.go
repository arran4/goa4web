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
	defer conn.Close()

	// Mock expectations for SchemaVersion
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM schema_version")).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(handlers.ExpectedSchemaVersion))

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
