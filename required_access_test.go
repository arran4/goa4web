package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func TestRequiredAccessAllowed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	queries := New(db)

	rows := sqlmock.NewRows([]string{"idpermissions", "users_idusers", "section", "level"}).
		AddRow(1, 1, "blogs", "writer")
	mock.ExpectQuery("SELECT idpermissions").WithArgs(int32(1), "blogs").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Idusers: 1})
	ctx = context.WithValue(ctx, ContextValues("queries"), queries)
	req = req.WithContext(ctx)

	if !RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	queries := New(db)

	rows := sqlmock.NewRows([]string{"idpermissions", "users_idusers", "section", "level"}).
		AddRow(1, 1, "blogs", "reader")
	mock.ExpectQuery("SELECT idpermissions").WithArgs(int32(1), "blogs").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Idusers: 1})
	ctx = context.WithValue(ctx, ContextValues("queries"), queries)
	req = req.WithContext(ctx)

	if RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
