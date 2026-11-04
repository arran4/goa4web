package searchworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	searchdomain "github.com/arran4/goa4web/internal/search"
)

// EventKey is retained as a compatibility alias for the search event key.
const EventKey = searchdomain.IndexEventKey

// Index types are retained as compatibility aliases for the search contract.
const (
	TypeComment = searchdomain.TypeComment
	TypeWriting = searchdomain.TypeWriting
	TypeLinker  = searchdomain.TypeLinker
	TypeImage   = searchdomain.TypeImage
)

// IndexEventData is retained as a compatibility alias for the search contract.
type IndexEventData = searchdomain.IndexEventData

// Worker listens for index events and updates the search tables.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier) {
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
			if !ok {
				env.Ack()
				continue
			}
			// Use a context that isn't canceled while processing
			// the current event so database operations aren't
			// interrupted mid-flight when the worker context is
			// cancelled.
			evtCtx := context.WithoutCancel(ctx)
			if data, ok := evt.Data[EventKey].(IndexEventData); ok {
				if err := index(evtCtx, q, data); err != nil {
					log.Printf("index error: %v", err)
				}
			} else {
				processEvent(evtCtx, evt, q)
			}
			env.Ack()
		case <-ctx.Done():
			return
		}
	}
}

func index(ctx context.Context, q db.Querier, data IndexEventData) error {
	return searchdomain.Index(ctx, q, data)
}

// Index synchronously updates the search index for one item.
func Index(ctx context.Context, q db.Querier, data IndexEventData) error {
	return index(ctx, q, data)
}
