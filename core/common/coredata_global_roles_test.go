package common

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

func TestAllRolesGlobalCaching(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}).
		AddRow(int32(1), "user", true, false, true, nil).
		AddRow(int32(2), "administrator", true, true, true, nil)

	// Expect query ONLY ONCE
	mock.ExpectQuery("SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles ORDER BY id").WillReturnRows(rows)

	// Create a shared cache
	cache := &lazy.Value[[]*db.Role]{}

	// Instance 1
	cd1 := NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig(), WithRolesCache(cache))
	roles1, err := cd1.AllRoles()
	if err != nil {
		t.Fatalf("AllRoles 1: %v", err)
	}
	if len(roles1) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles1))
	}

	// Instance 2
	cd2 := NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig(), WithRolesCache(cache))
	roles2, err := cd2.AllRoles()
	if err != nil {
		t.Fatalf("AllRoles 2: %v", err)
	}
	if len(roles2) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles2))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
