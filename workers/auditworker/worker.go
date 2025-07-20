package auditworker

import (
	"context"
	"log"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// Worker records bus events into the audit_log table.
func Worker(ctx context.Context, bus *eventbus.Bus, q *dbpkg.Queries) {
	if q == nil || bus == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case msg := <-ch:
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				continue
			}
			named, ok := evt.Task.(tasks.Name)
			if evt.UserID == 0 || !ok {
				continue
			}
			if admin, ok := evt.Task.(tasks.AdminTask); !ok || !admin.IsAdminTask() {
				continue
			}
			if err := q.InsertAuditLog(ctx, dbpkg.InsertAuditLogParams{
				UsersIdusers: evt.UserID,
				Action:       named.Name(),
			}); err != nil {
				log.Printf("insert audit log: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
