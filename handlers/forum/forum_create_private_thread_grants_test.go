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

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCreatePrivateThreadDoesNotSynthesizeParticipantReplyGrant(t *testing.T) {
	const (
		creatorID     int32 = 1
		participantID int32 = 2
		topicID       int32 = 5
		threadID      int32 = 100
	)

	queries := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	queries.SystemCreateThreadReturns = int64(threadID)
	queries.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
		Idforumtopic: topicID,
		Title:        sql.NullString{String: "Private topic", Valid: true},
		Handler:      "private",
	}
	queries.ListPrivateTopicParticipantsByTopicIDForUserReturns = []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
		{Idusers: participantID, Username: sql.NullString{String: "participant", Valid: true}},
	}
	queries.CreateCommentInSectionForCommenterResult = 999

	originalStore := core.Store
	originalSessionName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = originalStore
		core.SessionName = originalSessionName
	}()

	session := testhelpers.Must(core.Store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName))
	session.Values["UID"] = creatorID
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithSession(session))
	cd.UserID = creatorID
	cd.ForumBasePath = "/private"

	form := url.Values{
		"replytext": {"First post"},
		"language":  {"1"},
	}
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://example.com/private/topic/%d/thread", topicID), strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request = mux.SetURLVars(request, map[string]string{"topic": fmt.Sprint(topicID)})
	request = request.WithContext(context.WithValue(request.Context(), consts.KeyCoreData, cd))

	result := CreateThreadTaskHandler.Action(httptest.NewRecorder(), request)
	if err, ok := result.(error); ok {
		t.Fatalf("creating private thread: %v", err)
	}
	if len(queries.SystemCopyPrivateTopicGrantsToThreadCalls) != 1 {
		t.Fatalf("topic grant copy calls = %d, want 1", len(queries.SystemCopyPrivateTopicGrantsToThreadCalls))
	}
	wantCopy := db.SystemCopyPrivateTopicGrantsToThreadParams{ThreadID: threadID, TopicID: topicID}
	if got := queries.SystemCopyPrivateTopicGrantsToThreadCalls[0]; got != wantCopy {
		t.Errorf("topic grant copy = %#v, want %#v", got, wantCopy)
	}

	for _, call := range queries.AdminCreateGrantCalls {
		if call.UserID.Valid && call.UserID.Int32 == participantID && call.Action == consts.PermissionActionReply.String() {
			t.Fatalf("private thread creation synthesized reply grant for view-only participant: %#v", call)
		}
	}
}
