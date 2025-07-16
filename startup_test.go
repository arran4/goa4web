package goa4web

import (
	"context"
	dbstart2 "github.com/arran4/goa4web/internal/app/dbstart"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

func TestEnsureSchemaVersionMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM schema_version")).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(hcommon.ExpectedSchemaVersion))

	if err := dbstart2.EnsureSchema(context.Background(), db); err != nil {
		t.Fatalf("ensureSchema: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEnsureSchemaVersionMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM schema_version")).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(hcommon.ExpectedSchemaVersion - 1))

	err = dbstart2.EnsureSchema(context.Background(), db)
	if err == nil {
		t.Fatalf("expected error")
	}
	expected := dbstart2.RenderSchemaMismatch(hcommon.ExpectedSchemaVersion-1, hcommon.ExpectedSchemaVersion)
	if err.Error() != expected {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
