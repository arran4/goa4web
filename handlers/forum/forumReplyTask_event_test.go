package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/lazy"
)

// Ensure reply task populates notification data so admin emails render correctly.
func TestForumReplyTaskEventData(t *testing.T) {
	uid := int32(1)
	topicID := int32(1)
	threadID := int32(2)

	qs := &db.QuerierStub{
		GetPermissionsByUserIDFn: func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
			return []*db.GetPermissionsByUserIDRow{}, nil
		},
		SystemGetUserByIDFn: func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  uid,
				Username: sql.NullString{String: "testuser", Valid: true},
				Email:    sql.NullString{String: "test@example.com", Valid: true},
			}, nil
		},
		CreateCommentInSectionForCommenterFn: func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return 5, nil
		},
		GetCommentByIdForUserFn: func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
			return &db.GetCommentByIdForUserRow{
				Idcomments: 5,
			}, nil
		},
		RemoveContentPrivateLabelFn: func(ctx context.Context, arg db.RemoveContentPrivateLabelParams) error {
			return nil
		},
		ClearUnreadContentPrivateLabelExceptUserFn: func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error {
			return nil
		},
		UpsertContentReadMarkerFn: func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error {
			return nil
		},
		AddContentPrivateLabelFn: func(ctx context.Context, arg db.AddContentPrivateLabelParams) error {
			return nil
		},
		GetThreadBySectionThreadIDForReplierFn: func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return &db.Forumthread{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		},
		SystemInsertDeadLetterFn: func(ctx context.Context, message string) error {
			return nil
		},
		GetThreadLastPosterAndPermsFn: func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
			return &db.GetThreadLastPosterAndPermsRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		},
		GetForumTopicByIdForUserFn: func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: topicID,
				Title:        sql.NullString{String: "t", Valid: true},
			}, nil
		},
	}

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = uid
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = uid

	thread := &db.GetThreadLastPosterAndPermsRow{Idforumthread: threadID, ForumtopicIdforumtopic: topicID}
	if _, err := cd.ForumThreadByID(threadID, lazy.Set(thread)); err != nil {
		t.Fatalf("set thread: %v", err)
	}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Title: sql.NullString{String: "t", Valid: true}}
	if _, err := cd.ForumTopicByID(topicID, lazy.Set(topic)); err != nil {
		t.Fatalf("set topic: %v", err)
	}
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"hi"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/forum/topic/1/thread/2/reply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})

	rr := httptest.NewRecorder()
	replyTask.Action(rr, req)

	if _, ok := evt.Data["Username"].(string); !ok {
		t.Fatalf("username not set: %+v", evt.Data)
	}
	// The #c1 vs #c5 issue is because the URL generation logic might be defaulting to 1 or similar if it can't find the comment ID properly in the event data map or somewhere else.
	// But `CreateCommentInSectionForCommenter` returns 5.
	// Let's print evt.Data to see what's in there.
	// t.Logf("Event Data: %+v", evt.Data)

	// Wait, the test failure says `got ...#c1`. This 1 is likely coming from something else.
	// Ah, I suspect `HandleThreadUpdated` uses `event.CommentID`.
	// `CreateForumCommentForCommenter` calls `CreateCommentInSectionForCommenter` which returns `id`.
	// Then it calls `HandleThreadUpdated` with `CommentID: int32(id)`.
	// So it should be 5.
	// Why is it 1?
	// Maybe `CreateCommentInSectionForCommenterFn` isn't being called?
	// Or `CreateCommentInSectionForCommenter` wrapper in `CoreData` isn't using the result correctly?

	// Let's verify `CreateCommentInSectionForCommenterFn` was called.
	if len(qs.CreateCommentInSectionForCommenterCalls) == 0 {
		t.Fatalf("CreateCommentInSectionForCommenterFn was not called")
	}

	if v, ok := evt.Data["CommentURL"].(string); !ok || v != "/forum/topic/1/thread/2#c5" {
		t.Fatalf("comment URL: got %v, want /forum/topic/1/thread/2#c5. Data: %+v", v, evt.Data)
	}
}
