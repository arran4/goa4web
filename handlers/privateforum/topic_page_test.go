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
)

func TestTopicPage_Prefix(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	mockQueries := &db.QuerierProxier{
		OverwrittenGetForumCategoryById: func(ctx context.Context, arg db.GetForumCategoryByIdParams) (*db.Forumcategory, error) {
			return &db.Forumcategory{}, nil
		},
		OverwrittenGetForumTopicByIdForUser: func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: 1,
				Title:        sql.NullString{String: "topic", Valid: true},
				Lastaddition: sql.NullTime{Time: time.Now(), Valid: true},
				Handler:      "private",
			}, nil
		},
		OverwrittenListPrivateTopicParticipantsByTopicIDForUser: func(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
			return []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
				{Idusers: 1, Username: sql.NullString{String: "Alice", Valid: true}},
			}, nil
		},
		OverwrittenGetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText: func(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
			return []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
				{
					Idforumthread:          1,
					Firstpost:              1,
					Lastposter:             1,
					ForumtopicIdforumtopic: 1,
					Comments:               sql.NullInt32{Int32: 0, Valid: true},
					Lastposterusername:     sql.NullString{String: "Bob", Valid: true},
					Lastposterid:           sql.NullInt32{Int32: 1, Valid: true},
					Firstpostusername:      sql.NullString{String: "Alice", Valid: true},
					Firstposttext:          sql.NullString{String: "hi", Valid: true},
				},
			}, nil
		},
		OverwrittenSystemCheckRoleGrant: func(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
			return 1, nil
		},
		OverwrittenGetAllForumCategories: func(ctx context.Context, arg db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
			return []*db.Forumcategory{}, nil
		},
		OverwrittenGetPreferenceForLister: func(ctx context.Context, listerID int32) (*db.Preference, error) {
			return &db.Preference{Timezone: sql.NullString{String: "UTC", Valid: true}}, nil
		},
		OverwrittenListContentPublicLabels: func(ctx context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		},
		OverwrittenListContentLabelStatus: func(ctx context.Context, arg db.ListContentLabelStatusParams) ([]*db.ListContentLabelStatusRow, error) {
			return []*db.ListContentLabelStatusRow{}, nil
		},
		OverwrittenListContentPrivateLabels: func(ctx context.Context, arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
			return []*db.ListContentPrivateLabelsRow{}, nil
		},
	}
	cd := common.NewCoreData(context.Background(), mockQueries, config.NewRuntimeConfig())

	req := httptest.NewRequest(http.MethodGet, "/private/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	cd.ForumBasePath = "/private"
	cd.SetCurrentSection("privateforum")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	TopicPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "/private/topic/1/thread") {
		t.Fatalf("expected private thread link, got %q", body)
	}
	if !strings.Contains(body, `<nav class="breadcrumbs"`) || !strings.Contains(body, `href="/private">Private</a>`) {
		t.Fatalf("expected private breadcrumb, got %q", body)
	}

}
