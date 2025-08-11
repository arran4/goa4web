package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"

	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func TestMarkTopicReadTaskRedirect(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	form := url.Values{}
	form.Set("redirect", "/private/topic/1")
	form.Set("task", string(TaskMarkTopicRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/private/topic/1/thread/51")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	res := MarkTopicReadTaskHandler.Action(rr, req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/3" {
		t.Fatalf("redirect %q, want /private/topic/1/thread/3", rdh.TargetURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
  }
}

func TestSetLabelsTaskUpdatesSpecialLabels(t *testing.T) {
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO forumtopic_private_labels")).
		WithArgs(int32(1), cd.UserID, "new", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO forumtopic_private_labels")).
		WithArgs(int32(1), cd.UserID, "unread", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})

	res := MarkTopicReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1" {
		t.Fatalf("expected redirect to /private/topic/1 got %s", rdh.TargetURL)

  }
}

func TestMarkTopicReadTaskAddsInverseLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	mock.ExpectQuery("SELECT .* FROM forumtopic_public_labels").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"forumtopic_idforumtopic", "label"}))
	mock.ExpectQuery("SELECT .* FROM content_label_status").
		WithArgs("forumtopic", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectQuery("SELECT .* FROM forumtopic_private_labels").
		WithArgs(int32(1), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"forumtopic_idforumtopic", "users_idusers", "label", "invert"}))
	mock.ExpectExec("INSERT IGNORE INTO forumtopic_private_labels").
		WithArgs(int32(1), int32(1), "new", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT IGNORE INTO forumtopic_private_labels").
		WithArgs(int32(1), int32(1), "unread", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{}
	form.Set("task", string(TaskSetLabels))
	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	setLabelsTask.Action(rr, req)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
