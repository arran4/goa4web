package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	"github.com/arran4/goa4web/internal/db"
)

func TestTopicsPage_ThreadLinks(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	queries := db.New(conn)

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/forum/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	mock.ExpectQuery("SELECT .* FROM forumcategory").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}))

	mock.ExpectQuery("SELECT t.* FROM forumtopic t").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "lastposterusername"}).
			AddRow(1, 0, 1, 0, "Topic", "", 0, 0, time.Now(), "", ""))

	mock.ExpectQuery("SELECT .* FROM forumthread").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "lastposterusername", "lastposterid", "firstpostusername", "firstpostwritten", "firstposttext"}).
			AddRow(1, 0, 0, 1, 0, time.Now(), false, "abc", 0, "abc", time.Now(), "first post"))

	w := httptest.NewRecorder()
	TopicsPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "href=\"/forum/topic/1/thread/1\"") {
		t.Fatalf("expected thread link, got %q", body)
	}
	if strings.Contains(body, "href=\"//forum") {
		t.Fatalf("unexpected double slash in link: %q", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
