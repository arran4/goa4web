package news

import (
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	dbtest "github.com/arran4/goa4web/internal/db"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "User Roles") {
		t.Errorf("admin should see user roles")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbtest.New(db)
	ctx := req.Context()
	cd = common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.SetRoles([]string{"content writer", "administrator"})
	CustomNewsIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "User Roles") {
		t.Errorf("content writer should not see user roles")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer should see add news")
	}

	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	cd.SetRoles([]string{"anonymous"})
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Roles") || common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("anonymous should not see admin items")
	}
}
