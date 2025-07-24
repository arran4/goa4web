package faq

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"
	"regexp"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestAskActionPage_InvalidForms(t *testing.T) {
	dbconn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

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
		ctx := req.Context()
		ctx = context.WithValue(ctx, consts.KeyCoreData, &common.CoreData{})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(askTask)(rr, req)
		if rr.Code != http.StatusOK {
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

	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.AdminEmails = "a@test"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO faq (question, users_idusers, language_idlanguage) VALUES (?, ?, ?)")).
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
	bus := eventbus.NewBus()
	q := db.New(dbconn)
	evt := &eventbus.TaskEvent{Path: "/faq/ask", Task: tasks.TaskString(TaskAsk), UserID: 1}
	cd := common.NewCoreData(req.Context(), q)
	cd.UserID = 1
	cd.SetEvent(evt)

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	handler := middleware.TaskEventMiddlewareWithBus(bus)(http.HandlerFunc(handlers.TaskHandler(askTask)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	evt = cd.Event()
	if evt.Path != "/admin/faq" {
		t.Fatalf("event %+v", evt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
