package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
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

func TestPermissionUserAllowTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		bus := eventbus.NewBus()

		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			if id != 2 {
				return nil, sql.ErrNoRows
			}
			return &db.SystemGetUserByIDRow{
				Idusers:                2,
				Email:                  sql.NullString{String: "bob@test", Valid: true},
				Username:               sql.NullString{String: "bob", Valid: true},
				PublicProfileEnabledAt: sql.NullTime{},
			}, nil
		}
		queries.SystemGetUserByUsernameFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
			if username.String != "bob" {
				return nil, sql.ErrNoRows
			}
			return &db.SystemGetUserByUsernameRow{Idusers: 2, Username: sql.NullString{String: "bob", Valid: true}}, nil
		}

		ch := bus.Subscribe(eventbus.TaskMessageType)

		form := url.Values{}
		form.Set("username", "bob")
		form.Set("role", "moderator")
		form.Set("task", string(TaskUserAllow))
		req := httptest.NewRequest("POST", "/admin/user/2/permissions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"user": "2"})
		cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
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
		case env := <-ch:
			env.Ack()
			e, ok := env.Msg.(eventbus.TaskEvent)
			if !ok {
				t.Fatalf("wrong message type %T", env.Msg)
			}
			if e.Data["Username"] != "bob" || e.Data["Permission"] != "moderator" {
				t.Fatalf("unexpected event data: %+v", e.Data)
			}
		case <-time.After(time.Second):
			t.Fatal("no event")
		}
		if len(queries.SystemCreateUserRoleCalls) != 1 {
			t.Fatalf("expected create user role, got %d", len(queries.SystemCreateUserRoleCalls))
		}
		if arg := queries.SystemCreateUserRoleCalls[0]; arg.UsersIdusers != 2 || arg.Name != "moderator" {
			t.Fatalf("unexpected user role: %#v", arg)
		}
	})
}
