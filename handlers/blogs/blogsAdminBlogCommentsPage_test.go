package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminBlogCommentsPage_UsesURLParam(t *testing.T) {
	blogID := 9
	q := testhelpers.NewQuerierStub()
	q.GetBlogEntryForListerByIDRow = &db.GetBlogEntryForListerByIDRow{
		Idblogs:       int32(blogID),
		ForumthreadID: sql.NullInt32{Int32: 1, Valid: true},
		UsersIdusers:  1,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Blog:          sql.NullString{String: "body", Valid: true},
		Written:       time.Now(),
		Timezone:      sql.NullString{String: time.Local.String(), Valid: true},
		Username:      sql.NullString{String: "user", Valid: true},
		Comments:      0,
		IsOwner:       true,
	}
	q.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{
		{Idcomments: 1},
	}
	q.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
		Idforumthread: 1,
	}

	req := httptest.NewRequest("GET", "/admin/blogs/blog/"+strconv.Itoa(blogID)+"/comments", nil)
	req = mux.SetURLVars(req, map[string]string{"blog": strconv.Itoa(blogID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()

	AdminBlogCommentsPage(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
