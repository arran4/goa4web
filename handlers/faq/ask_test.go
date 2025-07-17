package faq

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestAskActionPage_InvalidForms(t *testing.T) {
	dbconn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	cases := []url.Values{
		{"text": {"hi"}},
		{"language": {"1"}},
		{"language": {"1"}, "text": {"hi"}, "foo": {"bar"}},
	}
	for _, form := range cases {
		req := httptest.NewRequest("POST", "/faq/ask", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
		ctx = context.WithValue(ctx, handlers.KeyCoreData, &handlers.CoreData{})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		AskActionPage(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("form=%v status=%d", form, rr.Code)
		}
	}
}

func TestAskActionPage_AdminEvent(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)

	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.AdminEmails = "a@test"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })

	mock.ExpectExec("INSERT INTO faq").
		WithArgs(sql.NullString{String: "hi", Valid: true}, int32(1), int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	form := url.Values{"language": {"1"}, "text": {"hi"}}
	req := httptest.NewRequest("POST", "/faq/ask", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	evt := &eventbus.Event{Path: "/faq/ask", Task: tasks.TaskAsk, UserID: 1}
	cd := &handlers.CoreData{}
	cd.SetEvent(evt)

	ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
	ctx = context.WithValue(ctx, handlers.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AskActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/faq" {
		t.Fatalf("location=%q", loc)
	}
	if !evt.Admin || evt.Path != "/admin/faq" {
		t.Fatalf("event %+v", evt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
