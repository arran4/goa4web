package backgroundtaskworker

import (
	"context"
	"log"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// Worker listens for task events implementing tasks.BackgroundTasker.
// The background method is executed and any returned task is published
// back onto the bus when the work completes.
func Worker(ctx context.Context, bus *eventbus.Bus, q *dbpkg.Queries) {
	if bus == nil || q == nil {
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
			if p, ok := evt.Task.(tasks.BackgroundTasker); ok {
				evtCtx := context.WithoutCancel(ctx)
				t, err := p.BackgroundTask(evtCtx, q)
				if err != nil {
					log.Printf("background task: %v", err)
					continue
				}
				if t != nil {
					nEvt := eventbus.TaskEvent{
						Path:   evt.Path,
						Task:   t,
						UserID: evt.UserID,
						Time:   time.Now(),
						Data:   evt.Data,
					}
					if err := bus.Publish(nEvt); err != nil && err != eventbus.ErrBusClosed {
						log.Printf("background publish: %v", err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
