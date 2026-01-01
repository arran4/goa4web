package privateforum

import (
	"context"
	"database/sql"
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

func TestTopicPage_Prefix(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	req := httptest.NewRequest(http.MethodGet, "/private/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	cd.ForumBasePath = "/private"
	cd.SetCurrentSection("privateforum")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	mock.ExpectQuery("SELECT .* FROM forumcategory").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}))

	mock.ExpectQuery("SELECT t.* FROM forumtopic t").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "lastposterusername"}).
			AddRow(1, 0, 0, 0, "topic", "", 0, 0, time.Now(), "private", ""))

	mock.ExpectQuery("SELECT u.idusers, u.username FROM grants").
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "username"}).AddRow(1, "Alice"))

	// GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText expects 3 args now:
	// viewer_id, topic_id, viewer_match_id
	// We match start of query "WITH role_ids AS"
	mock.ExpectQuery(`^-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many`).
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "lastposterusername", "lastposterid", "firstpostusername", "firstpostuserid", "firstpostwritten", "firstposttext"}).
			AddRow(1, 1, 1, 1, sql.NullInt32{Int32: 0, Valid: true}, sql.NullTime{}, sql.NullBool{}, sql.NullString{String: "Bob", Valid: true}, sql.NullInt32{Int32: 1, Valid: true}, sql.NullString{String: "Alice", Valid: true}, sql.NullInt32{Int32: 1, Valid: true}, sql.NullTime{}, sql.NullString{String: "hi", Valid: true}))

	w := httptest.NewRecorder()
	TopicPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "?error=") {
		t.Fatalf("page rendered with error: %s", body)
	}

	if !strings.Contains(body, "/private/topic/1/thread") {
		t.Fatalf("expected private thread link, got %q", body)
	}
	if !strings.Contains(body, `<nav class="breadcrumbs"`) || !strings.Contains(body, `href="/private">Private</a>`) {
		t.Fatalf("expected private breadcrumb, got %q", body)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
