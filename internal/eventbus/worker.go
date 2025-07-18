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
			name := ""
			if n, ok := evt.Task.(tasks.Name); ok {
				name = n.Name()
			}
			log.Printf("event path=%s task=%s uid=%d data=%v", evt.Path, name, evt.UserID, evt.Data)
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
			if evt.UserID == 0 || evt.Task == nil || !evt.Admin {
				continue
			}
			name := ""
			if n, ok := evt.Task.(tasks.Name); ok {
				name = n.Name()
			}
			if err := q.InsertAuditLog(ctx, dbpkg.InsertAuditLogParams{
				UsersIdusers: evt.UserID,
				Action:       name,
			}); err != nil {
				log.Printf("insert audit log: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
