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
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestPermissionUserTasksTemplates(t *testing.T) {
	admins := []notifications.AdminEmailTemplateProvider{
		&PermissionUserAllowTask{TaskString: TaskUserAllow},
		&PermissionUserDisallowTask{TaskString: TaskUserDisallow},
	}
	for _, p := range admins {
		et := p.AdminEmailTemplate()
		if et == nil || et.Text == "" || et.HTML == "" || et.Subject == "" {
			t.Errorf("incomplete templates for %T", p)
		}
		nt := p.AdminInternalNotificationTemplate()
		if nt == nil || *nt == "" {
			t.Errorf("missing internal template for %T", p)
		}
	}
}

func TestPermissionUserAllowEventData(t *testing.T) {
	bus := eventbus.NewBus()
	eventbus.DefaultBus = bus
	defer func() { eventbus.DefaultBus = eventbus.NewBus() }()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT idusers").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(2, "bob@test", "bob"))
	mock.ExpectExec("INSERT INTO user_roles").
		WithArgs(int32(2), "moderator").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ch := bus.Subscribe(eventbus.TaskMessageType)

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("role", "moderator")
	form.Set("task", string(TaskUserAllow))
	req := httptest.NewRequest("POST", "/admin/users/permissions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := &common.CoreData{}
	evt := &eventbus.TaskEvent{}
	cd.SetEvent(evt)
	ctx := req.Context()
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler := middleware.TaskEventMiddleware(http.HandlerFunc(tasks.Action(permissionUserAllowTask)))
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
