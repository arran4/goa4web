package logworker

import (
	"context"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/eventbus"
)

// Worker listens on the bus and logs all received events.
func Worker(ctx context.Context, bus *eventbus.Bus) {
	if bus == nil {
		return
	}
	ch, ack := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				ack()
				continue
			}
			log.Printf("event path=%s task=%s uid=%d data=%v", evt.Path, evt.Task, evt.UserID, cleanData(evt.Data))
			ack()
		case <-ctx.Done():
			return
		}
	}
}

func cleanData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	newData := make(map[string]any, len(data))
	for k, v := range data {
		if s, ok := v.(string); ok {
			s = strings.ReplaceAll(s, "\n", " ")
			s = strings.ReplaceAll(s, "\r", " ")
			if len(s) > 50 {
				runes := 0
				idx := -1
				for i := range s {
					if runes >= 50 {
						idx = i
						break
					}
					runes++
				}
				if idx != -1 {
					newData[k] = s[:idx] + "..."
				} else {
					newData[k] = s
				}
			} else {
				newData[k] = s
			}
		} else {
			newData[k] = v
		}
	}
	return newData
}
