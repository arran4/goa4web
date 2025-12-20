package faq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestAskActionPage_InvalidForms(t *testing.T) {
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
	queries := testutil.NewFAQQuerier(t)
	queries.AllowGrants()
	cfg := config.NewRuntimeConfig()

	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.AdminEmails = "a@test"
	cfg.EmailFrom = "from@example.com"
	cfg.NotificationsEnabled = true

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
	if len(queries.CreatedFAQArgs) != 0 {
		t.Fatalf("expected no create calls before action")
	}
	evt := &eventbus.TaskEvent{Path: "/faq/ask", Task: tasks.TaskString(TaskAsk), UserID: 1}
	cd := common.NewCoreData(req.Context(), queries, cfg, common.WithUserRoles([]string{"user"}), common.WithPreference(&db.Preference{PageSize: 15}))
	cd.UserID = 1
	cd.SetEvent(evt)

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mw := middleware.NewTaskEventMiddleware(bus)
	handler := mw.Middleware(http.HandlerFunc(handlers.TaskHandler(askTask)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	_ = cd.Event()
	if len(queries.CreatedFAQArgs) != 1 {
		t.Fatalf("expected one create call, got %d", len(queries.CreatedFAQArgs))
	}
	if queries.CreatedFAQArgs[0].Question.String != "hi" || queries.CreatedFAQArgs[0].WriterID != 1 {
		t.Fatalf("CreateFAQQuestionForWriter args %+v", queries.CreatedFAQArgs[0])
	}
}
