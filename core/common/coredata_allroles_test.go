package common

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestAllRolesLazy(t *testing.T) {
	queries := testutil.NewRolesQuerier(t)
	queries.Roles = []*db.Role{
		{ID: 1, Name: "user", CanLogin: true, IsAdmin: false, PrivateLabels: true},
		{ID: 2, Name: "administrator", CanLogin: true, IsAdmin: true, PrivateLabels: true},
	}

	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

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

}
