package eventbus

import (
	"context"
	"log"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// LogWorker listens on the bus and logs all received events.
func LogWorker(ctx context.Context, bus *Bus) {
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			log.Printf("event path=%s task=%s uid=%d data=%v", evt.Path, evt.Task, evt.UserID, evt.Data)
		case <-ctx.Done():
			return
		}
	}
}

// AuditWorker records bus events into the audit_log table.
func AuditWorker(ctx context.Context, bus *Bus, q *dbpkg.Queries) {
	if q == nil || bus == nil {
		return
	}
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
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
