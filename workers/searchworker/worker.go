package searchworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/searchindex"
)

// EventKey is retained as a compatibility alias for the searchindex event key.
const EventKey = searchindex.IndexEventKey

// Index types are retained as compatibility aliases for the searchindex contract.
const (
	TypeComment = searchindex.TypeComment
	TypeWriting = searchindex.TypeWriting
	TypeLinker  = searchindex.TypeLinker
	TypeImage   = searchindex.TypeImage
)

// IndexEventData is retained as a compatibility alias for the searchindex contract.
type IndexEventData = searchindex.IndexEventData

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
			// The server keeps the worker context alive while accepted events drain,
			// then cancels it at the shutdown deadline.
			evtCtx := ctx
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
	return searchindex.Index(ctx, q, data)
}

// Index synchronously updates the search index for one item.
func Index(ctx context.Context, q db.Querier, data IndexEventData) error {
	return index(ctx, q, data)
}
