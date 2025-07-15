package common

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestCoreDataLanguages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en")
	mock.ExpectQuery("SELECT idlanguage").WillReturnRows(rows)

	cd := NewCoreData(context.Background(), q)
	l, err := cd.Languages()
	if err != nil || len(l) != 1 {
		t.Fatalf("Languages: %v len=%d", err, len(l))
	}
	l2, err := cd.Languages()
	if err != nil || len(l2) != 1 {
		t.Fatalf("Languages second: %v len=%d", err, len(l2))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCoreDataUserRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "role"}).AddRow(1, 1, "admin")
	mock.ExpectQuery("SELECT ur.iduser_roles").WillReturnRows(rows)

	cd := NewCoreData(context.Background(), q)
	r, err := cd.UserRoles()
	if err != nil || len(r) != 1 {
		t.Fatalf("UserRoles: %v len=%d", err, len(r))
	}
	r2, err := cd.UserRoles()
	if err != nil || len(r2) != 1 {
		t.Fatalf("UserRoles second: %v len=%d", err, len(r2))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
