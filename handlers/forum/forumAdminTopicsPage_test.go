package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicsPage(t *testing.T) {
	queries := &db.QuerierStub{
		AdminCountForumTopicsReturn: 2,
		AdminListForumTopicsReturns: []*db.Forumtopic{
			{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 10,
				Title:                        sql.NullString{String: "Public topic", Valid: true},
				Threads:                      sql.NullInt32{Int32: 3, Valid: true},
				Comments:                     sql.NullInt32{Int32: 4, Valid: true},
			},
			{
				Idforumtopic:                 2,
				ForumcategoryIdforumcategory: 20,
				Title:                        sql.NullString{String: "Private topic", Valid: true},
				Threads:                      sql.NullInt32{Int32: 1, Valid: true},
				Comments:                     sql.NullInt32{Int32: 2, Valid: true},
				Handler:                      "private",
			},
		},
		GetAllForumCategoriesReturns: []*db.Forumcategory{
			{Idforumcategory: 10, Title: sql.NullString{String: "General", Valid: true}},
		},
		AdminGetTopicGrantsReturns: map[int32][]*db.AdminGetTopicGrantsRow{
			1: {
				{RoleName: sql.NullString{String: "anyone", Valid: true}},
			},
			2: {
				{Username: sql.NullString{String: "alice", Valid: true}},
				{Username: sql.NullString{String: "bob", Valid: true}},
			},
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

	cfg := config.NewRuntimeConfig()
	cfg.PageSizeDefault = 2
	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithOffset(4))
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
	if queries.AdminCountForumTopicsCalls != 1 {
		t.Fatalf("expected AdminCountForumTopics to be called once, got %d", queries.AdminCountForumTopicsCalls)
	}
	if got := len(queries.AdminListForumTopicsParams); got != 1 {
		t.Fatalf("AdminListForumTopics called %d times", got)
	}
	params := queries.AdminListForumTopicsParams[0]
	expectedLimit := int32(cd.PageSize())
	if params.Limit != expectedLimit || params.Offset != int32(cd.Offset()) {
		t.Fatalf("AdminListForumTopics params = %+v want limit=%d offset=%d", params, expectedLimit, cd.Offset())
	}
	if got := len(queries.GetAllForumCategoriesParams); got != 1 {
		t.Fatalf("GetAllForumCategories called %d times", got)
	}
	if queries.GetAllForumCategoriesParams[0].ViewerID != 0 {
		t.Fatalf("GetAllForumCategories viewer ID = %d", queries.GetAllForumCategoriesParams[0].ViewerID)
	}
	if got := len(queries.AdminGetTopicGrantsParams); got != len(queries.AdminListForumTopicsReturns) {
		t.Fatalf("AdminGetTopicGrants called %d times", got)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Public (Anyone)") {
		t.Fatalf("expected public access info in response")
	}
	if !strings.Contains(body, "alice, bob") {
		t.Fatalf("expected participants in response")
	}
}
