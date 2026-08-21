package forum

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

const (
	testForkUserID    int32 = 7
	testForkTopicID   int32 = 5
	testForkThreadID  int32 = 50
	testForkCommentID int32 = 60
	testNewThreadID   int32 = 100
)

func TestCreateForkRejectsInvalidSourceBeforeThreadInsert(t *testing.T) {
	tests := []struct {
		name       string
		quoteID    string
		configure  func(*db.QuerierStub)
		wantStatus int
	}{
		{name: "malformed", quoteID: "garbage", wantStatus: http.StatusBadRequest},
		{
			name:       "inaccessible comment",
			quoteID:    fmt.Sprint(testForkCommentID),
			wantStatus: http.StatusForbidden,
			configure: func(q *db.QuerierStub) {
				q.GetCommentByIdForUserErr = sql.ErrNoRows
			},
		},
		{
			name:       "cross topic",
			quoteID:    fmt.Sprint(testForkCommentID),
			wantStatus: http.StatusBadRequest,
			configure: func(q *db.QuerierStub) {
				q.GetThreadLastPosterAndPermsForUserReturns.ForumtopicIdforumtopic = 6
				q.GetForumTopicByIdForUserFn = func(_ context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
					return &db.GetForumTopicByIdForUserRow{Idforumtopic: arg.Idforumtopic}, nil
				}
			},
		},
		{
			name:       "comment thread mismatch",
			quoteID:    fmt.Sprint(testForkCommentID),
			wantStatus: http.StatusForbidden,
			configure: func(q *db.QuerierStub) {
				q.GetThreadLastPosterAndPermsForUserReturns.Idforumthread = testForkThreadID + 1
			},
		},
		{
			name:       "source is not replyable",
			quoteID:    fmt.Sprint(testForkCommentID),
			wantStatus: http.StatusForbidden,
			configure: func(q *db.QuerierStub) {
				q.GetThreadBySectionThreadIDForReplierErr = sql.ErrNoRows
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, request, recorder := newForkActionRequest(t, "", tt.quoteID)
			if tt.configure != nil {
				tt.configure(queries)
			}
			CreateThreadTaskHandler.Action(recorder, request)
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if len(queries.SystemCreateThreadCalls) != 0 || len(queries.SystemCreateReplyThreadCalls) != 0 {
				t.Fatalf("thread inserted before validation: ordinary=%v fork=%v", queries.SystemCreateThreadCalls, queries.SystemCreateReplyThreadCalls)
			}
		})
	}
}

func TestCreatePublicForkPersistsRelationshipInInitialInsert(t *testing.T) {
	queries, request, recorder := newForkActionRequest(t, "", fmt.Sprint(testForkCommentID))
	result := CreateThreadTaskHandler.Action(recorder, request)
	if err, ok := result.(error); ok {
		t.Fatalf("create public fork: %v", err)
	}
	if len(queries.SystemCreateReplyThreadCalls) != 1 || len(queries.SystemCreateThreadCalls) != 0 {
		t.Fatalf("create calls: fork=%v ordinary=%v", queries.SystemCreateReplyThreadCalls, queries.SystemCreateThreadCalls)
	}
	created := queries.SystemCreateReplyThreadCalls[0]
	if created.TopicID != testForkTopicID || created.ReplyToThreadID.Int32 != testForkThreadID || created.ReplyToCommentID.Int32 != testForkCommentID {
		t.Errorf("fork relationship = %#v", created)
	}
}

func TestCreatePrivateForkCopiesOnlySourceThreadGrants(t *testing.T) {
	queries, request, recorder := newForkActionRequest(t, "private", fmt.Sprint(testForkCommentID))
	result := CreateThreadTaskHandler.Action(recorder, request)
	if err, ok := result.(error); ok {
		t.Fatalf("create private fork: %v", err)
	}
	if len(queries.SystemCopyPrivateThreadGrantsToThreadCalls) != 1 {
		t.Fatalf("source thread grant copies = %d, want 1", len(queries.SystemCopyPrivateThreadGrantsToThreadCalls))
	}
	copyCall := queries.SystemCopyPrivateThreadGrantsToThreadCalls[0]
	if copyCall.SrcThreadID.Int32 != testForkThreadID || copyCall.DstThreadID.Int32 != testNewThreadID {
		t.Errorf("source grant copy = %#v", copyCall)
	}
	if len(queries.SystemCopyPrivateTopicGrantsToThreadCalls) != 0 {
		t.Fatalf("private fork broadened grants from topic: %#v", queries.SystemCopyPrivateTopicGrantsToThreadCalls)
	}
}

func TestCreateForkOpeningCommentFailureCleansUninitializedRelationship(t *testing.T) {
	queries, request, recorder := newForkActionRequest(t, "", fmt.Sprint(testForkCommentID))
	queries.CreateCommentInSectionForCommenterErr = errors.New("insert comment failed")
	result := CreateThreadTaskHandler.Action(recorder, request)
	if _, ok := result.(error); !ok {
		t.Fatalf("result = %#v, want error", result)
	}
	if len(queries.SystemCreateReplyThreadCalls) != 1 {
		t.Fatalf("fork inserts = %d, want 1", len(queries.SystemCreateReplyThreadCalls))
	}
	if len(queries.SystemDeleteUninitializedThreadCalls) != 1 || queries.SystemDeleteUninitializedThreadCalls[0] != testNewThreadID {
		t.Fatalf("cleanup calls = %#v, want thread %d", queries.SystemDeleteUninitializedThreadCalls, testNewThreadID)
	}
}

func newForkActionRequest(t *testing.T, handler, quoteID string) (*db.QuerierStub, *http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	queries := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	queries.SystemCreateReplyThreadReturns = int64(testNewThreadID)
	queries.SystemCreateThreadReturns = int64(testNewThreadID)
	queries.CreateCommentInSectionForCommenterResult = 999
	queries.GetForumTopicByIdForUserReturns = &db.GetForumTopicByIdForUserRow{
		Idforumtopic: testForkTopicID,
		Title:        sql.NullString{String: "Topic", Valid: true},
		Handler:      handler,
	}
	queries.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments:    testForkCommentID,
		ForumthreadID: testForkThreadID,
		Text:          sql.NullString{String: "source", Valid: true},
	}
	queries.GetThreadLastPosterAndPermsForUserReturns = &db.GetThreadLastPosterAndPermsForUserRow{
		Idforumthread:          testForkThreadID,
		ForumtopicIdforumtopic: testForkTopicID,
	}
	queries.GetThreadBySectionThreadIDForReplierReturn = &db.Forumthread{Idforumthread: testForkThreadID}

	store := sessions.NewCookieStore([]byte("fork-test"))
	session := sessions.NewSession(store, "fork-test")
	session.Values["UID"] = testForkUserID
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithSession(session))
	cd.UserID = testForkUserID
	if handler == "private" {
		cd.ForumBasePath = "/private"
	} else {
		cd.ForumBasePath = "/forum"
	}

	form := url.Values{"replytext": {"opening post"}, "language": {"1"}}
	requestURL := fmt.Sprintf("http://example.com%s/topic/%d/thread?quote_comment_id=%s", cd.ForumBasePath, testForkTopicID, url.QueryEscape(quoteID))
	request := httptest.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request = mux.SetURLVars(request, map[string]string{"topic": fmt.Sprint(testForkTopicID)})
	request = request.WithContext(context.WithValue(request.Context(), consts.KeyCoreData, cd))
	return queries, request, httptest.NewRecorder()
}
