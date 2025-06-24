package goa4web

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTemplateOverride(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	mock.ExpectExec("INSERT INTO template_overrides").WithArgs("t", "body").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.SetTemplateOverride(context.Background(), "t", "body"); err != nil {
		t.Fatalf("set: %v", err)
	}

	rows := sqlmock.NewRows([]string{"body"}).AddRow("body")
	mock.ExpectQuery("SELECT body FROM template_overrides").WithArgs("t").WillReturnRows(rows)
	body, err := q.GetTemplateOverride(context.Background(), "t")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if body != "body" {
		t.Fatalf("got %q", body)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
