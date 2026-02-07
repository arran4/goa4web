package admin

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestInterfaceDrivenBreadcrumbs(t *testing.T) {
	// Setup CoreData
	req := httptest.NewRequest("GET", "/admin/roles", nil)
	cfg := config.NewRuntimeConfig()
	queries := &BreadcrumbTestQuerier{}
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Test AdminRolesPage (Page implementation)
	rolesPage := &AdminRolesPage{}

	cd.SetCurrentPage(rolesPage)

	crumbs := cd.Breadcrumbs()

	// AdminRolesPage breadcrumb is "Roles" -> "Admin"
	// So we expect [Admin, Roles]

	if len(crumbs) != 2 {
		t.Fatalf("Expected 2 crumbs, got %d: %v", len(crumbs), crumbs)
	}

	if crumbs[0].Title != "Admin" || crumbs[0].Link != "/admin" {
		t.Errorf("First crumb mismatch: %v", crumbs[0])
	}
	if crumbs[1].Title != "Roles" || crumbs[1].Link != "/admin/roles" {
		t.Errorf("Second crumb mismatch: %v", crumbs[1])
	}
}

func TestDynamicBreadcrumb(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/role/1", nil)
	cfg := config.NewRuntimeConfig()
	queries := &BreadcrumbTestQuerier{}
	cd := common.NewCoreData(req.Context(), queries, cfg)

	page := &AdminRolePage{
		RoleName: "Moderator",
		RoleID:   1,
	}
	cd.SetCurrentPage(page)

	crumbs := cd.Breadcrumbs()
	// Expected: Admin -> Roles -> Role Moderator

	if len(crumbs) != 3 {
		t.Fatalf("Expected 3 crumbs, got %d: %v", len(crumbs), crumbs)
	}

	if crumbs[0].Title != "Admin" {
		t.Errorf("1st crumb wrong")
	}
	if crumbs[1].Title != "Roles" {
		t.Errorf("2nd crumb wrong")
	}
	if crumbs[2].Title != "Role Moderator" {
		t.Errorf("3rd crumb wrong: %s", crumbs[2].Title)
	}
}
