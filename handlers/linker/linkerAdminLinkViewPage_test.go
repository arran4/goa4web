package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminLinkViewPage(t *testing.T) {
	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow{
			ID:          1,
			LanguageID:  sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:    2,
			CategoryID:  sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:    0,
			Title:       sql.NullString{String: "t", Valid: true},
			Url:         sql.NullString{String: "http://u", Valid: true},
			Description: sql.NullString{String: "d", Valid: true},
			Listed:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Username:    sql.NullString{String: "bob", Valid: true},
			Title_2:     sql.NullString{String: "cat", Valid: true},
		},
	}
	dir := t.TempDir()
	siteDir := filepath.Join(dir, "site")
	if err := os.Mkdir(siteDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "adminLinkViewPage.gohtml"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	templates.SetDir(dir)
	t.Cleanup(func() { templates.SetDir("") })
	req := httptest.NewRequest("GET", "/admin/linker/links/link/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	adminLinkViewPage(w, req)

	if len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls) != 1 {
		t.Fatalf("expected one call, got %d", len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls))
	}
	if queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls[0] != 1 {
		t.Fatalf("unexpected link id %d", queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls[0])
	}
}
