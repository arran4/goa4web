package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestTopicPage_Prefix(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	queries := testhelpers.NewQuerierStub()
	queries.GetAllForumCategoriesReturns = []*db.Forumcategory{}
	queries.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: 1,
			Title:        sql.NullString{String: "topic", Valid: true},
			Handler:      "private",
			Lastaddition: sql.NullTime{Time: time.Now(), Valid: true},
		}, nil
	}
	queries.ListPrivateTopicParticipantsByTopicIDForUserReturns = []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
		{Idusers: 1, Username: sql.NullString{String: "Alice", Valid: true}},
		{Idusers: 2, Username: sql.NullString{String: "Bob", Valid: true}},
	}
	queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn = func(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
		return []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
			{
				Idforumthread:          1,
				Firstpost:              1,
				Lastposter:             1,
				ForumtopicIdforumtopic: 1,
				Comments:               sql.NullInt32{Int32: 0, Valid: true},
				Lastaddition:           sql.NullTime{},
				Locked:                 sql.NullBool{},
				Lastposterusername:     sql.NullString{String: "Bob", Valid: true},
				Lastposterid:           sql.NullInt32{Int32: 1, Valid: true},
				Firstpostusername:      sql.NullString{String: "Alice", Valid: true},
				Firstpostuserid:        sql.NullInt32{Int32: 1, Valid: true},
				Firstpostwritten:       sql.NullTime{},
				Firstposttext:          sql.NullString{String: "hi", Valid: true},
			},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest(http.MethodGet, "/private/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	cd.ForumBasePath = "/private"
	cd.SetCurrentSection("privateforum")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	TopicPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "?error=") {
		t.Fatalf("page rendered with error: %s", body)
	}

	if !strings.Contains(body, "/private/topic/1/thread") {
		t.Fatalf("expected private thread link, got %q", body)
	}
	if !strings.Contains(body, `<nav class="breadcrumbs"`) || !strings.Contains(body, `href="/private">Private</a>`) {
		t.Fatalf("expected private breadcrumb, got %q", body)
	}
}
