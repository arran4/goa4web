package dbstart

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/handlers"
)

func TestEnsureSchemaVersionMatch(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"version_id", "is_applied"}).AddRow(handlers.ExpectedSchemaVersion, true))

	if err := EnsureSchema(context.Background(), conn); err != nil {
		t.Fatalf("ensureSchema: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEnsureSchemaVersionMismatch(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT version_id, is_applied from goose_db_version ORDER BY id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"version_id", "is_applied"}).AddRow(handlers.ExpectedSchemaVersion-1, true))

	err = EnsureSchema(context.Background(), conn)
	if err == nil {
		t.Fatalf("expected error")
	}
	expected := RenderSchemaMismatch(handlers.ExpectedSchemaVersion-1, handlers.ExpectedSchemaVersion)
	if err.Error() != expected {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
