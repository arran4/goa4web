package common

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestAllRolesLazy(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}).
		AddRow(int32(1), "user", true, false, true, nil).
		AddRow(int32(2), "administrator", true, true, true, nil)

	mock.ExpectQuery("SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles ORDER BY id").WillReturnRows(rows)

	cd := NewTestCoreData(t, db.New(conn))

	roles, err := cd.AllRoles()
	if err != nil {
		t.Fatalf("AllRoles: %v", err)
	}
	if len(roles) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles))
	}
	if roles[0].Name != "anyone" || roles[0].ID != 0 {
		t.Fatalf("expected anyone role first, got %+v", roles[0])
	}
	if _, err := cd.AllRoles(); err != nil {
		t.Fatalf("AllRoles second call: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
