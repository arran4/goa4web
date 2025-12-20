package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	uid := int32(1)
	topicID := int32(1)
	threadID := int32(2)

	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at").
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(uid, "test@example.com", "testuser", nil))

	mock.ExpectExec("INSERT INTO comments").
		WithArgs(int32(1), uid, threadID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), topicID, sqlmock.AnyArg(), uid).
		WillReturnResult(sqlmock.NewResult(5, 1))

	mock.ExpectExec("DELETE FROM content_private_labels").
		WithArgs("thread", threadID, uid).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO content_read_markers").
		WithArgs("thread", threadID, uid, 5).
		WillReturnResult(sqlmock.NewResult(1, 1))

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = uid
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
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
	if v, ok := evt.Data["CommentURL"].(string); !ok || v != "/forum/topic/1/thread/2#c1" {
		t.Fatalf("comment URL: %+v", evt.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
