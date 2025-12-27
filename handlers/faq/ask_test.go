package faq

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
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
	q := &db.QuerierStub{}
	cd := common.NewCoreData(req.Context(), q, cfg)
	cd.UserID = 1

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mw := middleware.NewTaskEventMiddleware(bus)
	handler := mw.Middleware(http.HandlerFunc(handlers.TaskHandler(askTask)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if got := len(q.SystemCheckGrantCalls); got != 1 {
		t.Fatalf("expected 1 grant check, got %d", got)
	}
	grant := q.SystemCheckGrantCalls[0]
	if grant.Section != "faq" || grant.Item.String != "question" || grant.Action != "post" {
		t.Fatalf("unexpected grant check: %#v", grant)
	}
	if grant.UserID.Int32 != 1 || !grant.UserID.Valid {
		t.Fatalf("unexpected grant user: %#v", grant.UserID)
	}
	if got := len(q.CreateFAQQuestionForWriterCalls); got != 1 {
		t.Fatalf("expected 1 faq insert, got %d", got)
	}
	createCall := q.CreateFAQQuestionForWriterCalls[0]
	if createCall.Question.String != "hi" || createCall.GranteeID.Int32 != 1 || createCall.WriterID != 1 {
		t.Fatalf("unexpected insert params: %#v", createCall)
	}
	if createCall.LanguageID.Int32 != 1 || !createCall.LanguageID.Valid {
		t.Fatalf("unexpected language params: %#v", createCall.LanguageID)
	}
	if evt := cd.Event(); evt == nil || evt.Path != "/admin/faq/questions" {
		t.Fatalf("unexpected event path: %#v", evt)
	}
}
