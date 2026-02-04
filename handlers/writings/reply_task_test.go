package writings

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/go-be-lazy"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
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
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("writing", "article", "reply"))
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	now := time.Now()
	writingID := int32(1)
	threadID := int32(3)
	topicID := int32(4)

	q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{Idwriting: writingID, ForumthreadID: threadID}
	q.SystemGetForumTopicByTitleReturns = &db.Forumtopic{
		Idforumtopic: topicID,
		Title:        sql.NullString{String: common.WritingTopicName, Valid: true},
		Lastaddition: sql.NullTime{Time: now, Valid: true},
		Handler:      "writing",
	}
	q.GetForumTopicByIdReturns = q.SystemGetForumTopicByTitleReturns
	q.CreateCommentInSectionForCommenterResult = 5
	q.ClearUnreadContentPrivateLabelExceptUserFn = func(context.Context, db.ClearUnreadContentPrivateLabelExceptUserParams) error {
		return nil
	}

	_, _ = cd.WritingByID(writingID, lazy.Set(q.GetWritingForListerByIDRow))
	cd.SetCurrentWriting(writingID)

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

	if len(q.CreateCommentInSectionForCommenterCalls) != 1 {
		t.Fatalf("expected 1 comment insert, got %d", len(q.CreateCommentInSectionForCommenterCalls))
	}
	if got := q.CreateCommentInSectionForCommenterCalls[0]; got.ForumthreadID != threadID || got.ItemID.Int32 != writingID {
		t.Fatalf("unexpected comment insert call: %+v", got)
	}
}
