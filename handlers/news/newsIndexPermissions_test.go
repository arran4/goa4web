package news

import (
	"context"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	corecommon "github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	ctx := context.WithValue(req.Context(), corecorecommon.KeyQueries, q)
	cd = corecommon.NewCoreData(ctx, q)
	cd.SetRoles([]string{"content writer", "administrator"})
	CustomNewsIndex(cd, req.WithContext(ctx))
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("content writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer should see add news")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("anonymous should not see admin items")
	}
}
