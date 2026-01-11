package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicsPage(t *testing.T) {
	queries := &db.QuerierStub{
		AdminCountForumTopicsReturns: 1,
		AdminListForumTopicsReturns: []*db.Forumtopic{
			{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "t", Valid: true},
				Description:                  sql.NullString{String: "d", Valid: true},
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
				Handler:                      "",
			},
		},
		GetAllForumCategoriesReturns: []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		},
		AdminGetTopicGrantsReturns: []*db.AdminGetTopicGrantsRow{
			{Section: "forum", RoleID: sql.NullInt32{}, RoleName: sql.NullString{}, UserID: sql.NullInt32{}, Username: sql.NullString{}},
		},
	}

	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics", nil)
	sess, _ := core.Store.New(r, core.SessionName)
	ctx := context.WithValue(r.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	AdminTopicsPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
}
