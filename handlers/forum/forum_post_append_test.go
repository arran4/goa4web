
package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForumPostAppend_Fallback(t *testing.T) {
	t.Run("Fallback to normal reply", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: 1, Handler: "forum", Title: sql.NullString{String: "Title", Valid: true},
			}, nil
		}
		q.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread: 2, ForumtopicIdforumtopic: 1,
			}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.ForumPostAppendWindow = 60

		req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/2/reply", nil)
		req.Form = url.Values{"replytext": {"new text"}, "language": {"1"}}
		req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})

		sess := &sessions.Session{Values: map[any]any{"UID": int32(5)}}
		cd := common.NewCoreData(context.Background(), q, cfg, common.WithSession(sess))
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		q.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return 11, nil
		}

		q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		    return 1, nil
		}

		replyTask.Action(httptest.NewRecorder(), req)

		if len(q.CreateCommentInSectionForCommenterCalls) != 1 {
			t.Fatalf("expected fallback to create comment, got %d calls", len(q.CreateCommentInSectionForCommenterCalls))
		}
	})
}
