package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminUserWritingsPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		userID := int32(1)
		qs := testhelpers.NewQuerierStub()
		qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
			Idusers:                userID,
			Email:                  sql.NullString{String: "u@test", Valid: true},
			Username:               sql.NullString{String: "user", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		}
		qs.AdminGetAllWritingsByAuthorReturns = []*db.AdminGetAllWritingsByAuthorRow{{
			Idwriting:         1,
			UsersIdusers:      userID,
			ForumthreadID:     0,
			LanguageID:        sql.NullInt32{},
			WritingCategoryID: 2,
			Title:             sql.NullString{String: "Title", Valid: true},
			Published:         sql.NullTime{Time: time.Now(), Valid: true},
			Timezone:          sql.NullString{String: time.Local.String(), Valid: true},
			Writing:           sql.NullString{String: "", Valid: true},
			Abstract:          sql.NullString{String: "", Valid: true},
			Private:           sql.NullBool{Bool: false, Valid: true},
			DeletedAt:         sql.NullTime{},
			LastIndex:         sql.NullTime{},
			Username:          sql.NullString{String: "user", Valid: true},
			Comments:          0,
		}}

		req := httptest.NewRequest("GET", "/admin/user/1/writings", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, qs, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
		cd.SetCurrentProfileUserID(userID)
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		adminUserWritingsPage(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
		body := rr.Body.String()
		if !strings.Contains(body, `<td>1</td>`) {
			t.Fatalf("missing id: %s", body)
		}
		if !strings.Contains(body, "Title") {
			t.Fatalf("missing title: %s", body)
		}
	})
}
