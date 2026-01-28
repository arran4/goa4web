package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
    "net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/lazy"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForumReplyRedirect(t *testing.T) {
	replierUID := int32(1)
	topicID := int32(5)
	threadID := int32(42)
    commentID := int64(999)

	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
        return &db.SystemGetUserByIDRow{
            Idusers:  replierUID,
            Username: sql.NullString{String: "replier", Valid: true},
        }, nil
	}
	qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return commentID, nil
	}
	qs.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
		return &db.Forumthread{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
		}, nil
	}
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
			Lastposterusername:     sql.NullString{String: "replier", Valid: true},
			Comments:               sql.NullInt32{Int32: 5, Valid: true}, // 5 comments before this one
		}, nil
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: topicID,
			Title:        sql.NullString{String: "Test Topic", Valid: true},
            Handler:      "forum",
		}, nil
	}

    // We mock GetCommentsByThreadIdForUser to return 6 comments (5 existing + 1 new)
    // This will be used when we fix the code.
    qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
        comments := make([]*db.GetCommentsByThreadIdForUserRow, 6)
        for i := 0; i < 6; i++ {
            comments[i] = &db.GetCommentsByThreadIdForUserRow{
                Idcomments: int32(100 + i), // Arbitrary IDs
            }
        }
        // Ensure one matches the new ID just in case
        comments[5].Idcomments = int32(commentID)
        return comments, nil
    }

    // Also mock permissions check used by thread loading
    qs.GetPermissionsByUserIDFn = func(id int32) ([]*db.GetPermissionsByUserIDRow, error) {
        return []*db.GetPermissionsByUserIDRow{}, nil
    }


	cfg := config.NewRuntimeConfig()

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = replierUID

	task := replyTask
	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  replierUID,
		Path:    "/forum/topic/5/thread/42/reply",
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = replierUID

	thread := &db.GetThreadLastPosterAndPermsRow{Idforumthread: threadID, ForumtopicIdforumtopic: topicID, Lastposterusername: sql.NullString{String: "replier", Valid: true}, Comments: sql.NullInt32{Int32: 5, Valid: true}}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Title: sql.NullString{String: "Test Topic", Valid: true}, Handler: "forum"}
	cd.SetCurrentThreadAndTopic(threadID, topicID)
	_, _ = cd.ForumThreadByID(threadID, lazy.Set(thread))
	_, _ = cd.ForumTopicByID(topicID, lazy.Set(topic))

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"Test reply"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/forum/topic/5/thread/42/reply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "5", "thread": "42"})

	rr := httptest.NewRecorder()
    result := task.Action(rr, req)

    redirect, ok := result.(handlers.RedirectHandler)
    if !ok {
        t.Fatalf("Expected RedirectHandler, got %T", result)
    }

    // Expectation: currently it redirects to #c6 (index)
    expectedUrl := "/forum/topic/5/thread/42#c6"
    if string(redirect) != expectedUrl {
        t.Errorf("Expected redirect to %q, got %q", expectedUrl, string(redirect))
    }
}
