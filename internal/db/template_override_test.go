package db

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
	if err := q.AdminSetTemplateOverride(context.Background(), AdminSetTemplateOverrideParams{Name: "t", Body: "body"}); err != nil {
		t.Fatalf("set: %v", err)
	}

	rows := sqlmock.NewRows([]string{"body"}).AddRow("body")
	mock.ExpectQuery("SELECT body FROM template_overrides").WithArgs("t").WillReturnRows(rows)
	body, err := q.SystemGetTemplateOverride(context.Background(), "t")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if body != "body" {
		t.Fatalf("got %q", body)
	}

	mock.ExpectExec("DELETE FROM template_overrides").WithArgs("t").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.AdminDeleteTemplateOverride(context.Background(), "t"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
