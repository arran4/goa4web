package logworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/eventbus"
)

// Worker listens on the bus and logs all received events.
func Worker(ctx context.Context, bus *eventbus.Bus) {
	if bus == nil {
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
			log.Printf("event path=%s task=%s uid=%d data=%v", evt.Path, evt.Task, evt.UserID, evt.Data)
		case <-ctx.Done():
			return
		}
	}
}
