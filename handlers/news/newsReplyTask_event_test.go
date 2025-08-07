package news

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// Ensure reply task populates notification data so admin emails render correctly.
func TestNewsReplyTaskEventData(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	uid := int32(1)
	pid := 2
	pthid := int32(3)

	mock.ExpectQuery("SELECT u.username AS writerName").
		WithArgs(uid, int32(pid), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"writerName", "writerId", "idsiteNews", "forumthread_id", "language_idlanguage", "users_idusers", "news", "occurred", "comments"}).
			AddRow("writer", uid, pid, pthid, 1, uid, "txt", time.Now(), 0))

	mock.ExpectQuery("SELECT idforumtopic").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_idlanguage", "title", "description", "threads", "comments", "lastaddition", "handler"}).
			AddRow(4, int32(0), 0, 0, NewsTopicName, "", 0, 0, sql.NullTime{}, "news"))

	mock.ExpectQuery("SELECT u.idusers").
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
			AddRow(uid, nil, "alice", nil))

	mock.ExpectExec("INSERT INTO comments").
		WithArgs(int32(1), uid, pthid, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), uid).
		WillReturnResult(sqlmock.NewResult(5, 1))

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = uid
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"administrator"}))
	cd.UserID = uid
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"hi"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/news/2", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"news": "2"})

	rr := httptest.NewRecorder()
	handlers.TaskHandler(replyTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if evt.Data["Username"] != "alice" {
		t.Fatalf("username not set: %+v", evt.Data)
	}
	if evt.Data["PostURL"] != "/news/news/2" {
		t.Fatalf("post url not set: %+v", evt.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
