package forum

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForumReplyErrorRetainsText(t *testing.T) {
	replierUID := int32(1)
	topicID := int32(5)
	threadID := int32(42)

	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		return &db.SystemGetUserByIDRow{
			Idusers:  replierUID,
			Username: sql.NullString{String: "replier", Valid: true},
		}, nil
	}
	qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 0, fmt.Errorf("simulated failure")
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Handler: "forum"}, nil
	}
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread: threadID,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{}, nil
	}
	qs.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
		return &db.Forumthread{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
		}, nil
	}
	qs.GetPermissionsByUserIDFn = func(id int32) ([]*db.GetPermissionsByUserIDRow, error) {
		return []*db.GetPermissionsByUserIDRow{
			{
				IsAdmin: true,
			},
		}, nil
	}
	qs.GetForumTopicByIdFn = func(ctx context.Context, id int32) (*db.Forumtopic, error) {
		return &db.Forumtopic{Idforumtopic: topicID, Handler: "forum"}, nil
	}

	cd := common.NewTestCoreData(t, qs)
	cd.UserID = replierUID
	cd.SetCurrentThreadAndTopic(threadID, topicID)
	cd.SetSession(sessions.NewSession(sessions.NewCookieStore([]byte("test")), "test"))

	formValues := url.Values{
		"replytext": []string{"This is my thoughtful reply"},
	}
	bodyStr := formValues.Encode()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/forum/topic/%d/thread/%d/reply", topicID, threadID), strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{
		"topicID":  fmt.Sprintf("%d", topicID),
		"threadID": fmt.Sprintf("%d", threadID),
	})

	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	oldHandle := tasks.Handle
	defer func() { tasks.Handle = oldHandle }()
	tasks.Handle = func(ww http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		s := fmt.Sprintf("%#v", data)
		if strings.Contains(s, "This is my thoughtful reply") {
			fmt.Fprintf(w, "This is my thoughtful reply")
		}
		if cd.CurrentError() != "" {
			fmt.Fprintf(w, "simulated failure")
		}
		return nil
	}

	ReplyTaskHandler.Action(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK after error, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "This is my thoughtful reply") {
		t.Errorf("expected reply text to be retained in output, got: %s", body)
	}
	if !strings.Contains(body, "simulated failure") {
		t.Errorf("expected error message in output, got: %s", body)
	}
}
