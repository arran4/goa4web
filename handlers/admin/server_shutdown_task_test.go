package admin

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathServerShutdownTask_EventPublished(t *testing.T) {
	bus := eventbus.NewBus()
	h := New(WithServer(&server.Server{Bus: bus}))
	ch := bus.Subscribe(eventbus.TaskMessageType)
	q := testhelpers.NewQuerierStub()

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(),
		common.WithUserRoles([]string{"administrator"}),
		common.WithPermissions([]*db.GetPermissionsByUserIDRow{
			{Name: "administrator", IsAdmin: true},
		}),
	)
	cd.UserID = 1
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.NewServerShutdownTask().Action(rr, req)

	select {
	case env := <-ch:
		env.Ack()
		evt, ok := env.Msg.(eventbus.TaskEvent)
		if !ok {
			t.Fatalf("wrong message type %T", env.Msg)
		}
		name, ok := evt.Task.(tasks.Name)
		if !ok {
			t.Fatalf("task does not implement Name")
		}
		if name.Name() != string(TaskServerShutdown) || evt.Path != "/admin/shutdown" || evt.UserID != 1 {
			t.Fatalf("unexpected event: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("event not published")
	}
}
