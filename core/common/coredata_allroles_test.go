package common

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestAllRolesLazy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin"}).
		AddRow(int32(1), "user", true, false).
		AddRow(int32(2), "administrator", true, true)

	mock.ExpectQuery("SELECT id, name, can_login, is_admin FROM roles ORDER BY id").WillReturnRows(rows)

	cd := NewCoreData(context.Background(), dbpkg.New(db))

	if _, err := cd.AllRoles(); err != nil {
		t.Fatalf("AllRoles: %v", err)
	}
	if _, err := cd.AllRoles(); err != nil {
		t.Fatalf("AllRoles second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
