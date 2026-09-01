package templates_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func renderForumReplyForm(t *testing.T, hasOtherRead bool) string {
	t.Helper()
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("forum", "topic", "append"))
	comment := &db.GetCommentsByThreadIdForUserRow{
		Idcomments: 20, ForumthreadID: 10, UsersIdusers: 7, IsOwner: true,
		Written: sql.NullTime{Time: time.Now(), Valid: true},
	}
	q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 5, Handler: "forum"}, nil
	}
	q.GetThreadBySectionThreadIDForReplierFn = func(context.Context, db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
		return &db.Forumthread{Idforumthread: 10}, nil
	}
	q.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{{
		Idcomments: comment.Idcomments, ForumthreadID: comment.ForumthreadID,
		UsersIdusers: comment.UsersIdusers, IsOwner: comment.IsOwner, Written: comment.Written,
	}}
	q.GetCommentsByThreadIdForUserFn = func(context.Context, db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{comment}, nil
	}
	q.SystemHasOtherUserReadItemAtOrBeyondFn = func(context.Context, db.SystemHasOtherUserReadItemAtOrBeyondParams) (bool, error) {
		return hasOtherRead, nil
	}
	cd := common.NewCoreData(context.Background(), q, &config.RuntimeConfig{ForumPostAppendWindow: 60}, common.WithUserRoles([]string{"user"}))
	cd.UserID = 7
	cd.SetCurrentSection("forum")
	cd.SetCurrentThreadAndTopic(10, 5)
	req := httptest.NewRequest("GET", "/forum/topic/5/thread/10", nil)
	compiled := templates.GetCompiledSiteTemplates(cd.Funcs(req), templates.WithSilence(true))
	data := map[string]any{
		"BasePath": "/forum",
		"Text":     "next segment",
		"Topic":    &db.GetForumTopicByIdForUserRow{Idforumtopic: 5, Handler: "forum"},
		"Thread":   &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 10},
	}
	var buf bytes.Buffer
	if err := compiled.ExecuteTemplate(&buf, "forumReply", data); err != nil {
		t.Fatalf("ExecuteTemplate forumReply: %v", err)
	}
	return buf.String()
}

func TestForumReplyFormAdvertisesAppendAdvisory(t *testing.T) {
	eligible := renderForumReplyForm(t, false)
	if !strings.Contains(eligible, `<span class="section-title">Append:</span>`) || !strings.Contains(eligible, `value="Reply">Append</button>`) {
		t.Fatalf("eligible form did not advertise append: %s", eligible)
	}
	read := renderForumReplyForm(t, true)
	if !strings.Contains(read, `<span class="section-title">Reply:</span>`) || !strings.Contains(read, `value="Reply">Reply</button>`) {
		t.Fatalf("read-marker-blocked form did not fall back to reply: %s", read)
	}
}
