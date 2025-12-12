package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/gorilla/mux"
)

func TestPermissionUserTasksTemplates(t *testing.T) {
	admins := []notifications.AdminEmailTemplateProvider{
		&PermissionUserAllowTask{TaskString: TaskUserAllow},
		&PermissionUserDisallowTask{TaskString: TaskUserDisallow},
	}
	for _, p := range admins {
		et, _ := p.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if et == nil || et.Text == "" || et.HTML == "" || et.Subject == "" {
			t.Errorf("incomplete templates for %T", p)
		}
		nt := p.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if nt == nil || *nt == "" {
			t.Errorf("missing internal template for %T", p)
		}
	}
}

func TestPermissionUserAllowEventData(t *testing.T) {
	bus := eventbus.NewBus()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	queries := db.New(conn)

	mock.ExpectQuery("FROM users").
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(2, "bob@test", "bob", nil))
	mock.ExpectQuery("FROM users").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).AddRow(2, "bob", nil))
	mock.ExpectExec("INSERT INTO user_roles").
		WithArgs(int32(2), "moderator").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ch := bus.Subscribe(eventbus.TaskMessageType)

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("role", "moderator")
	form.Set("task", string(TaskUserAllow))
	req := httptest.NewRequest("POST", "/admin/user/2/permissions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"user": "2"})
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	cd.LoadSelectionsFromRequest(req)
	evt := &eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}
	cd.SetEvent(evt)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	mw := middleware.NewTaskEventMiddleware(bus)
	handler := mw.Middleware(http.HandlerFunc(handlers.TaskHandler(permissionUserAllowTask)))
	handler.ServeHTTP(rr, req)

	select {
	case msg := <-ch:
		e, ok := msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong message type %T", msg)
		}
		if e.Data["Username"] != "bob" || e.Data["Permission"] != "moderator" {
			t.Fatalf("unexpected event data: %+v", e.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("no event")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
