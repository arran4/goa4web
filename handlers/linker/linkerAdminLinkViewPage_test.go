package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminLinkViewPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		req := httptest.NewRequest("GET", "/admin/linker/links/link/1", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "1"})
		w := httptest.NewRecorder()

		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow{
			ID:          1,
			LanguageID:  sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:    2,
			CategoryID:  sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:    0,
			Title:       sql.NullString{String: "t", Valid: true},
			Url:         sql.NullString{String: "http://u", Valid: true},
			Description: sql.NullString{String: "d", Valid: true},
			Listed:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Timezone:    sql.NullString{String: "UTC", Valid: true},
			Username:    sql.NullString{String: "bob", Valid: true},
			Title_2:     sql.NullString{String: "cat", Valid: true},
		}

		adminLinkViewPage(w, req)
	})
}
