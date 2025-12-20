package user

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/gorilla/mux"
)

type permissionQueries struct {
	db.Querier
	userID     int32
	user       *db.SystemGetUserByIDRow
	username   string
	userByName *db.SystemGetUserByUsernameRow
	created    []db.SystemCreateUserRoleParams
}

func (q *permissionQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *permissionQueries) SystemGetUserByUsername(_ context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
	if username.String != q.username {
		return nil, fmt.Errorf("unexpected username: %s", username.String)
	}
	return q.userByName, nil
}

func (q *permissionQueries) SystemCreateUserRole(_ context.Context, arg db.SystemCreateUserRoleParams) error {
	q.created = append(q.created, arg)
	return nil
}

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

	queries := &permissionQueries{
		userID:   2,
		username: "bob",
		user: &db.SystemGetUserByIDRow{
			Idusers:                2,
			Email:                  sql.NullString{String: "bob@test", Valid: true},
			Username:               sql.NullString{String: "bob", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
		userByName: &db.SystemGetUserByUsernameRow{Idusers: 2, Username: "bob"},
	}

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
	if len(queries.created) != 1 {
		t.Fatalf("expected create user role, got %d", len(queries.created))
	}
	if arg := queries.created[0]; arg.UsersIdusers != 2 || arg.Name != "moderator" {
		t.Fatalf("unexpected user role: %#v", arg)
	}
}
