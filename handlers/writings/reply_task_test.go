package writings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestReplyTaskAutoSubscribe(t *testing.T) {
	var task ReplyTask
	if _, ok := interface{}(task).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("AutoSubscribeProvider must auto subscribe as users will want updates")
	}
}

func TestReplyMarksWritingUnread(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	now := time.Now()
	writingID := int32(1)
	threadID := int32(3)

	_, _ = cd.WritingByID(writingID, lazy.Set(&db.GetWritingForListerByIDRow{Idwriting: writingID, ForumthreadID: threadID}))
	cd.SetCurrentWriting(writingID)

	mock.ExpectQuery("(?s).*SELECT 1 FROM grants.*").
		WithArgs(cd.UserID, "writing", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("SELECT idforumtopic").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler"}).
			AddRow(1, 0, 0, 1, common.WritingTopicName, "desc", 0, 0, now, "writing"))

	mock.ExpectExec("INSERT INTO comments").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), threadID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "writing", sqlmock.AnyArg(), writingID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(5, 1))

	mock.ExpectExec("DELETE FROM content_private_labels").
		WithArgs("writing", writingID, cd.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{}
	form.Set("replytext", "hi")
	form.Set("language", "1")
	req := httptest.NewRequest(http.MethodPost, "/writings/article/1/reply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"writing": "1"})

	sess := &sessions.Session{}
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	replyTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
