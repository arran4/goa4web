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
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicDeleteConfirmPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	// Point to the root directory for templates so they can be found if not using embedded
	cd.Config.TemplatesDir = "../../core/templates"

	r := httptest.NewRequest("GET", "/admin/forum/topics/topic/123/delete", nil)
	r = mux.SetURLVars(r, map[string]string{"topic": "123"})

	sess, _ := core.Store.New(r, core.SessionName)
	ctx := context.WithValue(r.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	AdminTopicDeleteConfirmPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Are you sure you want to delete forum topic 123?") {
		t.Errorf("expected confirmation message in body")
	}
	if !strings.Contains(body, "Confirm delete") {
		t.Errorf("expected confirm button label in body")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestAdminTopicDeletePage_NoCascade(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectExec("DELETE FROM forumtopic WHERE idforumtopic = ?").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())

	form := url.Values{}
	form.Add("task", "Confirm delete")
	// No cascade param

	r := httptest.NewRequest("POST", "/admin/forum/topics/topic/123/delete", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = mux.SetURLVars(r, map[string]string{"topic": "123"})

	ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	AdminTopicDeletePage(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status: got %d want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/admin/forum/topics" {
		t.Errorf("location: got %s want /admin/forum/topics", loc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestAdminTopicDeletePage_WithCascade(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	// Update expectation to match the actual complex query
	mock.ExpectExec(regexp.QuoteMeta("DELETE forumthread, comments, comments_search FROM forumthread LEFT JOIN comments ON comments.forumthread_id = forumthread.idforumthread LEFT JOIN comments_search ON comments_search.comment_id = comments.idcomments WHERE forumthread.forumtopic_idforumtopic = ?")).
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 5))

	mock.ExpectExec("DELETE FROM forumtopic WHERE idforumtopic = ?").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())

	form := url.Values{}
	form.Add("task", "Confirm delete")
	form.Add("cascade", "true")

	r := httptest.NewRequest("POST", "/admin/forum/topics/topic/123/delete", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = mux.SetURLVars(r, map[string]string{"topic": "123"})

	ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	AdminTopicDeletePage(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status: got %d want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/admin/forum/topics" {
		t.Errorf("location: got %s want /admin/forum/topics", loc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
