package news

import (
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	ctx := req.Context()
	cd = common.NewCoreData(ctx, q)
	cd.SetRoles([]string{"content writer", "administrator"})
	CustomNewsIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("content writer should not see user permissions")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer should see add news")
	}

	cd = common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Permissions") || common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("anonymous should not see admin items")
	}
}
