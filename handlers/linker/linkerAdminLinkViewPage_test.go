package linker

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminLinkViewPage(t *testing.T) {
	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingReturns: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow{
			ID:          1,
			Title:       sql.NullString{String: "t", Valid: true},
			Url:         sql.NullString{String: "http://u", Valid: true},
			Description: sql.NullString{String: "d", Valid: true},
		},
	}
	req := httptest.NewRequest("GET", "/admin/linker/links/link/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	adminLinkViewPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, w.Code)
	}
	if calls := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls; len(calls) != 1 {
		t.Fatalf("expected 1 link fetch call got %d", len(calls))
	} else if calls[0] != 1 {
		t.Fatalf("expected link id 1 got %d", calls[0])
	}
}
