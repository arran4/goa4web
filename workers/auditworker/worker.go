package auditworker

import (
	"context"
	"database/sql"
	"encoding/json"
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
			aud, ok := evt.Task.(tasks.AuditableTask)
			if !ok {
				continue
			}
			details := aud.AuditRecord(evt.Data)
			data, _ := json.Marshal(evt.Data)
			if err := q.SystemInsertAuditLog(ctx, dbpkg.SystemInsertAuditLogParams{
				UsersIdusers: evt.UserID,
				Action:       named.Name(),
				Path:         evt.Path,
				Details:      sql.NullString{String: details, Valid: details != ""},
				Data:         sql.NullString{String: string(data), Valid: len(data) > 0},
			}); err != nil {
				log.Printf("insert audit log: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
