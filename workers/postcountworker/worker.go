package postcountworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// EventKey is the map key used for post count update events.
const EventKey = "post_count"

// UpdateEventData describes which comment, thread and topic counts to refresh.
type UpdateEventData struct {
	ThreadID  int32
	TopicID   int32
	CommentID int32
}

// Worker listens for post count events and updates the related metadata.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier) {
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
			data, ok := evt.Data[EventKey].(UpdateEventData)
			if ok {
				if err := PostUpdate(ctx, q, data.ThreadID, data.TopicID); err != nil {
					log.Printf("post count update: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
