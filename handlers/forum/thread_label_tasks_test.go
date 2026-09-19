package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestMarkThreadReadTaskRedirect(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/2")
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/2" {
		t.Fatalf("redirect %q, want /private/topic/1/thread/2", rdh.TargetURL)
	}
}

func TestMarkThreadReadTaskRefererFallback(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/private/topic/1/thread/1")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/1" {
		t.Fatalf("redirect %q, want /private/topic/1/thread/1", rdh.TargetURL)
	}
}

func TestSetLabelsTaskAddsInverseLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT item, item_id, label\\s+FROM content_public_labels\\s+WHERE item = \\? AND item_id = \\?").
		WithArgs("thread", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectQuery("SELECT item, item_id, label\\s+FROM content_label_status\\s+WHERE item = \\? AND item_id = \\?").
		WithArgs("thread", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectQuery("SELECT item, item_id, user_id, label, invert\\s+FROM content_private_labels\\s+WHERE item = \\? AND item_id = \\? AND user_id = \\?").
		WithArgs("thread", int32(1), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO content_private_labels")).
		WithArgs("thread", int32(1), int32(1), "new", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO content_private_labels")).
		WithArgs("thread", int32(1), int32(1), "unread", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{}
	form.Set("task", string(TaskSetLabels))
	req := httptest.NewRequest(http.MethodPost, "/forum/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	setLabelsTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSetLabelsTaskUpdatesSpecialLabels(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
		}, nil
	}

	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())
	cd.UserID = 2

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(TaskMarkThreadRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	_ = MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)

	// verify that AddContentPrivateLabel was called for 'new' and 'unread' with Invert=true
	foundNew := false
	foundUnread := false
	for _, call := range qs.AddContentPrivateLabelCalls {
		if call.Item == "thread" && call.ItemID == 1 && call.UserID == cd.UserID && call.Invert == true {
			if call.Label == "new" {
				foundNew = true
			}
			if call.Label == "unread" {
				foundUnread = true
			}
		}
	}
	if !foundNew || !foundUnread {
		t.Fatalf("expected inverted new and unread labels to be added, got new: %v, unread: %v", foundNew, foundUnread)
	}
	if !foundNew || !foundUnread {
		t.Fatalf("expected new and unread labels to be removed, got new: %v, unread: %v", foundNew, foundUnread)
	}
}

func TestMarkThreadReadTaskRedirectWithThread(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(TaskMarkThreadRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/3" {
		t.Fatalf("expected redirect to /private/topic/1/thread/3 got %s", rdh.TargetURL)
	}
}

func TestMarkThreadReadTaskExplicitLastComment(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/2")
	form.Set("task", string(TaskMarkThreadRead))
	form.Set("last_comment", "101")
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/2" {
		t.Fatalf("expected redirect to /private/topic/1/thread/2 got %s", rdh.TargetURL)
	}

	if len(qs.UpsertContentReadMarkerCalls) != 1 {
		t.Fatalf("expected 1 UpsertContentReadMarker, got %d", len(qs.UpsertContentReadMarkerCalls))
	}
	marker := qs.UpsertContentReadMarkerCalls[0]
	if marker.LastCommentID != 101 {
		t.Fatalf("expected explicit LastCommentID to be 101, got %d", marker.LastCommentID)
	}
}

func TestMarkThreadReadTaskFallbackLastComment(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/2")
	form.Set("task", string(TaskMarkThreadRead))
	// last_comment is omitted entirely
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/2" {
		t.Fatalf("expected redirect to /private/topic/1/thread/2 got %s", rdh.TargetURL)
	}

	if len(qs.UpsertContentReadMarkerCalls) != 1 {
		t.Fatalf("expected 1 UpsertContentReadMarker, got %d", len(qs.UpsertContentReadMarkerCalls))
	}
	marker := qs.UpsertContentReadMarkerCalls[0]
	if marker.LastCommentID != 102 {
		t.Fatalf("expected fallback LastCommentID to be 102, got %d", marker.LastCommentID)
	}
}

func TestMarkThreadReadTaskInvalidLastComment(t *testing.T) {
	qs := testhelpers.NewQuerierStub(testhelpers.WithDefaultGrantAllowed(true))
	qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
		}, nil
	}
	qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 101},
			{Idcomments: 102},
		}, nil
	}
	cd := common.NewCoreData(context.Background(), qs, config.NewRuntimeConfig())

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/2")
	form.Set("task", string(TaskMarkThreadRead))
	form.Set("last_comment", "invalid")
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/2" {
		t.Fatalf("expected redirect to /private/topic/1/thread/2 got %s", rdh.TargetURL)
	}

	if len(qs.UpsertContentReadMarkerCalls) != 1 {
		t.Fatalf("expected 1 UpsertContentReadMarker, got %d", len(qs.UpsertContentReadMarkerCalls))
	}
	marker := qs.UpsertContentReadMarkerCalls[0]
	if marker.LastCommentID != 102 {
		t.Fatalf("expected fallback LastCommentID to be 102 for invalid input, got %d", marker.LastCommentID)
	}
}
