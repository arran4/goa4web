package notifications

import (
	"context"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// DatabaseFeederWorker polls the notifications table and publishes lightweight
// feed events for new rows. This keeps the worker role as a DB-driven feeder.
func (n *Notifier) DatabaseFeederWorker(ctx context.Context, interval time.Duration) {
	if n == nil || n.Queries == nil || n.Bus == nil {
		return
	}

	lastID := int32(0)
	if rows, err := n.Queries.AdminListRecentNotifications(ctx, 1); err == nil && len(rows) > 0 {
		lastID = rows[0].ID
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := n.Queries.AdminListRecentNotifications(ctx, 100)
			if err != nil {
				log.Printf("notification feeder list recent: %v", err)
				continue
			}
			// rows are newest-first; publish oldest-first to preserve order.
			for i := len(rows) - 1; i >= 0; i-- {
				row := rows[i]
				if row.ID <= lastID {
					continue
				}
				evt := eventbus.TaskEvent{
					Path:   "/admin/notifications",
					Task:   tasks.TaskString("NotificationFeed"),
					UserID: row.UsersIdusers,
					Time:   row.CreatedAt,
					Data: map[string]any{
						"notificationID": row.ID,
					},
					Outcome: eventbus.TaskOutcomeSuccess,
				}
				if err := n.Bus.Publish(evt); err != nil && err != eventbus.ErrBusClosed {
					log.Printf("notification feeder publish: %v", err)
				}
				lastID = row.ID
			}
		}
	}
}
