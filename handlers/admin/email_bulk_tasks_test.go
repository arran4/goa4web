package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type bulkEmailQueries struct {
	db.QuerierStub
	pendingEmails       map[int32]*db.AdminGetPendingEmailByIDRow
	sentIDs             []int32
	failedIDs           []int32
	markSentCalls       []int32
	listSentIDsCalled   bool
	listFailedIDsCalled bool
	getPendingCalls     []int32
}

func (q *bulkEmailQueries) AdminListSentEmailIDs(_ context.Context, _ db.AdminListSentEmailIDsParams) ([]int32, error) {
	q.listSentIDsCalled = true
	return q.sentIDs, nil
}

func (q *bulkEmailQueries) AdminListFailedEmailIDs(_ context.Context, _ db.AdminListFailedEmailIDsParams) ([]int32, error) {
	q.listFailedIDsCalled = true
	return q.failedIDs, nil
}

func (q *bulkEmailQueries) AdminGetPendingEmailByID(_ context.Context, id int32) (*db.AdminGetPendingEmailByIDRow, error) {
	q.getPendingCalls = append(q.getPendingCalls, id)
	if q.pendingEmails[id] == nil {
		return nil, sql.ErrNoRows
	}
	return q.pendingEmails[id], nil
}

func (q *bulkEmailQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	return &db.SystemGetUserByIDRow{
		Idusers: id,
		Username: sql.NullString{
			String: "test-user",
			Valid:  true,
		},
		Email: sql.NullString{
			String: "user@example.com",
			Valid:  true,
		},
	}, nil
}

func (q *bulkEmailQueries) SystemMarkPendingEmailSent(_ context.Context, id int32) error {
	q.markSentCalls = append(q.markSentCalls, id)
	return nil
}

func setupBulkEmailTaskRequest(t *testing.T, target string, form url.Values, queries *bulkEmailQueries) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	return req.WithContext(ctx)
}

func TestHappyPathResendSentEmailTaskFilteredSelection(t *testing.T) {
	queries := &bulkEmailQueries{
		pendingEmails: map[int32]*db.AdminGetPendingEmailByIDRow{
			11: {
				ID:       11,
				ToUserID: sql.NullInt32{Int32: 5, Valid: true},
				Body:     "To: user@example.com\r\nSubject: Hello\r\n\r\nBody",
			},
			12: {
				ID:       12,
				ToUserID: sql.NullInt32{Int32: 5, Valid: true},
				Body:     "To: user@example.com\r\nSubject: Hello\r\n\r\nBody",
			},
		},
		sentIDs: []int32{11, 12},
	}
	form := url.Values{}
	form.Set("selection", "filtered")
	form.Set("task", string(TaskResend))
	req := setupBulkEmailTaskRequest(t, "/admin/email/sent?role=moderator", form, queries)
	rr := httptest.NewRecorder()

	result := resendSentEmailTask.Action(rr, req)
	rdh, ok := result.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", result)
	}
	if !strings.Contains(rdh.TargetURL, "status=resent") || !strings.Contains(rdh.TargetURL, "scope=filtered") {
		t.Fatalf("expected redirect to include bulk status, got %s", rdh.TargetURL)
	}
	if !queries.listSentIDsCalled {
		t.Fatal("expected sent email id list query to be called")
	}
}

func TestHappyPathResendQueueTaskFilteredFailedSelection(t *testing.T) {
	queries := &bulkEmailQueries{
		pendingEmails: map[int32]*db.AdminGetPendingEmailByIDRow{
			31: {
				ID:       31,
				ToUserID: sql.NullInt32{Int32: 7, Valid: true},
				Body:     "To: user@example.com\r\nSubject: Failed\r\n\r\nBody",
			},
		},
		failedIDs: []int32{31},
	}
	form := url.Values{}
	form.Set("selection", "filtered")
	form.Set("task", string(TaskResend))
	req := setupBulkEmailTaskRequest(t, "/admin/email/failed?lang=1", form, queries)
	rr := httptest.NewRecorder()

	result := resendQueueTask.Action(rr, req)
	rdh, ok := result.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", result)
	}
	if !strings.Contains(rdh.TargetURL, "status=resent") || !strings.Contains(rdh.TargetURL, "scope=filtered") {
		t.Fatalf("expected redirect to include bulk status, got %s", rdh.TargetURL)
	}
	if !queries.listFailedIDsCalled {
		t.Fatal("expected failed email id list query to be called")
	}
	if len(queries.markSentCalls) != 1 || queries.markSentCalls[0] != 31 {
		t.Fatalf("expected sent mark for id 31, got %v", queries.markSentCalls)
	}
}
