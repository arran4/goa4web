package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func TestMarkThreadReadTaskRedirect(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
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
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
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
	defer conn.Close()
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
	form.Set("task", string(forumcommon.TaskSetLabels))
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO content_private_labels")).
		WithArgs("thread", int32(1), cd.UserID, "new", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO content_private_labels")).
		WithArgs("thread", int32(1), cd.UserID, "unread", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(forumcommon.TaskMarkThreadRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	// Execute the mark-as-read task, which should upsert the inverse labels.
	_ = MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMarkThreadReadTaskRedirectWithThread(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(forumcommon.TaskMarkThreadRead))
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
