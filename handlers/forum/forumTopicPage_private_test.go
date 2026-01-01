package forum

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

func TestTopicsPage_PrivateTopic(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"

	qs := &db.QuerierStub{
		GetAllForumCategoriesFn: func(ctx context.Context, arg db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
			return []*db.Forumcategory{}, nil
		},
		GetForumTopicByIdForUserFn: func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "Topic: old is now Private chat with: Bob", Valid: true},
				Handler:                      "private",
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		},
		ListPrivateTopicParticipantsByTopicIDForUserFn: func(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
			return []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
				{Idusers: 1, Username: sql.NullString{String: "Alice", Valid: true}},
				{Idusers: 2, Username: sql.NullString{String: "Bob", Valid: true}},
			}, nil
		},
		GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn: func(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
			return []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{}, nil
		},
		ListContentPublicLabelsFn: func(arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		},
	}

	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())
	cd.UserID = 1 // Set viewer ID to 1 (Alice)

	req := httptest.NewRequest(http.MethodGet, "/forum/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	TopicsPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "is now Private chat with") {
		t.Fatalf("unexpected conversion message in output: %q", body)
	}
	if strings.Contains(body, "Category:") {
		t.Fatalf("unexpected category heading: %q", body)
	}
	if !strings.Contains(body, "Topic: Bob") {
		t.Fatalf("expected participant names (Bob, as Alice is viewer), got %q", body)
	}
}
