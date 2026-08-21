package backgroundtaskworker

import (
	"context"
	"log"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// Worker listens for task events implementing tasks.BackgroundTasker.
// The background method is executed and any returned task is published
// back onto the bus when the work completes.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier, cfg *config.RuntimeConfig) {
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
			if p, ok := evt.Task.(tasks.BackgroundTasker); ok {
				evtCtx := context.WithoutCancel(ctx)
				cd := common.NewCoreData(evtCtx, q, cfg)
				evtCtx = context.WithValue(evtCtx, consts.KeyCoreData, cd)
				go func() {
					defer env.Ack()
					t, err := p.BackgroundTask(evtCtx, q)
					if err != nil {
						log.Printf("background task: %v", err)
						return
					}
					if t != nil {
						nEvt := eventbus.TaskEvent{
							Path:    evt.Path,
							Task:    t,
							UserID:  evt.UserID,
							Time:    time.Now(),
							Data:    evt.Data,
							Outcome: eventbus.TaskOutcomeSuccess,
						}
						if err := bus.Publish(nEvt); err != nil && err != eventbus.ErrBusClosed {
							log.Printf("background publish: %v", err)
						}
					}
				}()
			} else {
				env.Ack()
			}
		case <-ctx.Done():
			return
		}
	}
}
