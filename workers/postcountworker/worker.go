package postcountworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/eventbus"
	forumdomain "github.com/arran4/goa4web/internal/forum"
)

// EventKey is retained as a compatibility alias for the forum event key.
const EventKey = forumdomain.PostCountEventKey

// UpdateEventData is retained as a compatibility alias for the forum contract.
type UpdateEventData = forumdomain.PostCountUpdate

// Worker listens for post count events and updates the related metadata.
func Worker(ctx context.Context, bus *eventbus.Bus, q PostUpdateQuerier) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case env, ok := <-ch:
			if !ok {
				return
			}
			evt, ok := env.Msg.(eventbus.TaskEvent)
			if ok {
				data, ok := evt.Data[EventKey].(UpdateEventData)
				if ok {
					if err := PostUpdate(ctx, q, data.ThreadID, data.TopicID); err != nil {
						log.Printf("post count update: %v", err)
					}
				}
			}
			env.Ack()
		case <-ctx.Done():
			return
		}
	}
}
