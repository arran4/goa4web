package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminCategoryCreateSubmitSuccess(t *testing.T) {
	queries := &db.QuerierStub{
		GetAllForumCategoriesReturns: []*db.Forumcategory{},
		AdminCreateForumCategoryFn: func(ctx context.Context, arg db.AdminCreateForumCategoryParams) (int64, error) {
			if arg.ParentID != 1 {
				t.Fatalf("unexpected parent id %d", arg.ParentID)
			}
			if arg.CategoryLanguageID != (sql.NullInt32{Int32: 2, Valid: true}) {
				t.Fatalf("unexpected language id %+v", arg.CategoryLanguageID)
			}
			if arg.Title.String != "name" || !arg.Title.Valid {
				t.Fatalf("unexpected title %+v", arg.Title)
			}
			if arg.Description.String != "desc" || !arg.Description.Valid {
				t.Fatalf("unexpected desc %+v", arg.Description)
			}
			return 5, nil
		},
	}
	form := url.Values{
		"name":     {"name"},
		"desc":     {"desc"},
		"pcid":     {"1"},
		"language": {"2"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/forum/categories/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoryCreateSubmit(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/admin/forum/categories") || !strings.Contains(location, "error=category created") {
		t.Fatalf("unexpected redirect location %q", location)
	}
}

func TestAdminCategoryCreateSubmitValidationError(t *testing.T) {
	queries := &db.QuerierStub{}
	form := url.Values{
		"desc":     {"desc"},
		"pcid":     {"1"},
		"language": {"2"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/forum/categories/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoryCreateSubmit(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/admin/forum/categories/create") || !strings.Contains(location, "error=missing name") {
		t.Fatalf("expected validation error in redirect, got %q", location)
	}
}
